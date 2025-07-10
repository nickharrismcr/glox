package compiler

import (
	"fmt"
	"strconv"

	"glox/src/core"
	debug "glox/src/debug"
)

type Precedence int

const (
	PREC_NONE       Precedence = iota
	PREC_ASSIGNMENT            // =
	PREC_OR                    // or
	PREC_AND                   // and
	PREC_EQUALITY              // == !=
	PREC_COMPARISON            // < > <= >=
	PREC_TERM                  // + -
	PREC_FACTOR                // * / %
	PREC_UNARY                 // ! -
	PREC_CALL                  // . () []
	PREC_PRIMARY
)

type ParseFn func(*Parser, bool)

type ParseRule struct {
	prefix ParseFn
	infix  ParseFn
	prec   Precedence
}

type Local struct {
	name       Token
	lexeme     string
	depth      int
	isCaptured bool
}

type Loop struct {
	start     int
	breaks    []int
	foreach   bool
	continue_ int
}

type Upvalue struct {
	index   uint8
	isLocal bool
}

// NewLoop creates and returns a new Loop structure for managing loop control flow.
// Loop structures track loop state including start position, break statements,
// foreach loop type, and continue jump positions for break/continue implementation.
func NewLoop() *Loop {

	return &Loop{}
}

type FunctionType int

const (
	TYPE_FUNCTION FunctionType = iota
	TYPE_SCRIPT
	TYPE_METHOD
	TYPE_INITIALIZER
)

type ClassCompiler struct {
	enclosing     *ClassCompiler
	hasSuperClass bool
}

type Compiler struct {
	enclosing   *Compiler
	function    *core.FunctionObject
	type_       FunctionType
	locals      [256]*Local
	localCount  int
	scopeDepth  int
	loop        *Loop
	upvalues    [256]*Upvalue
	scriptName  string
	environment *core.Environment
}

type Name struct {
	Token Token
	Str   string
}

// NewCompiler creates and initializes a new compiler instance for compiling Lox functions.
// It sets up the compiler with the specified function type (script, function, method, or initializer),
// script name for debugging, parent compiler for nested scopes, and environment for module context.
// The compiler manages local variables, upvalues, scope depth, and generates bytecode for the function.
// Slot 0 is reserved for the enclosing function context ("this" for methods, empty for functions).
func NewCompiler(type_ FunctionType, scriptName string, parent *Compiler, environment *core.Environment) *Compiler {

	rv := &Compiler{
		enclosing:   parent,
		type_:       type_,
		scriptName:  scriptName,
		function:    core.MakeFunctionObject(scriptName, environment),
		environment: environment,
	}
	// slot 0 is for enclosing function
	rv.locals[0] = &Local{
		depth:      0,
		isCaptured: false,
	}
	if type_ != TYPE_FUNCTION {
		rv.locals[0].name = SyntheticToken("this")
	} else {
		rv.locals[0].name = Token{}
	}
	rv.localCount = 1

	chunk := rv.function.Chunk
	chunk.LocalVars = append(chunk.LocalVars, core.LocalVarInfo{
		Name:    "",
		StartIp: 0,
		EndIp:   -1,
		Slot:    0,
	})
	return rv
}

type Parser struct {
	scn                 *Scanner
	current, previous   Token
	hadError, panicMode bool
	rules               map[TokenType]ParseRule
	currentCompiler     *Compiler
	currentClass        *ClassCompiler
	globals             map[string]bool
}

// NewParser creates and initializes a new parser instance for parsing Lox source code.
// The parser maintains state for current and previous tokens, error handling flags,
// parsing rules for different token types, current compiler context, class context,
// and tracks global variable declarations to prevent redefinition errors.
func NewParser() *Parser {

	p := &Parser{
		hadError:  false,
		panicMode: false,
		globals:   map[string]bool{},
	}
	p.setRules()
	return p
}

// Compile is the main entry point for compiling Lox source code into bytecode.
// It takes the script filename, source code string, and module name, then:
// 1. Creates a new parser and scanner for the source
// 2. Sets up a new compiler with TYPE_SCRIPT for top-level execution
// 3. Parses all declarations until EOF is reached
// 4. Returns the compiled function object containing bytecode, or nil if compilation failed
// Debug tracing can be enabled to monitor compilation progress.
func Compile(script string, source string, module string) *core.FunctionObject {

	if core.DebugTraceExecution && !core.DebugSuppress {
		fmt.Printf("Compiling %s\n", script)
	}
	parser := NewParser()

	parser.scn = NewScanner(source)
	environment := core.NewEnvironment(module)
	parser.currentCompiler = NewCompiler(TYPE_SCRIPT, script, nil, environment)
	parser.advance()
	for !parser.match(TOKEN_EOF) {
		parser.declaration()
	}
	//parser.consume(TOKEN_EOF, "Expect end of expression")
	function := parser.endCompiler()
	if core.DebugTraceExecution && !core.DebugSuppress {
		fmt.Println("Compile done.")
	}
	if parser.hadError {
		return nil
	}
	return function
}

// setRules initializes the parsing rules table that maps token types to their
// corresponding prefix/infix parsing functions and operator precedence levels.
// This implements Pratt parsing (top-down operator precedence parsing) where:
// - prefix functions handle tokens that appear at the start of expressions
// - infix functions handle binary operators and postfix operations
// - precedence determines the order of operations for expression parsing
func (p *Parser) setRules() {

	p.rules = map[TokenType]ParseRule{
		TOKEN_LEFT_PAREN:    {prefix: grouping, infix: call, prec: PREC_CALL},
		TOKEN_RIGHT_PAREN:   {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_LEFT_BRACE:    {prefix: dictLiteral, infix: nil, prec: PREC_NONE},
		TOKEN_RIGHT_BRACE:   {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_LEFT_BRACKET:  {prefix: listLiteral, infix: slice, prec: PREC_CALL},
		TOKEN_RIGHT_BRACKET: {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_COMMA:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_DOT:           {prefix: nil, infix: dot, prec: PREC_CALL},
		TOKEN_MINUS:         {prefix: unary, infix: binary, prec: PREC_TERM},
		TOKEN_PLUS:          {prefix: nil, infix: binary, prec: PREC_TERM},
		TOKEN_PLUS_PLUS:     {prefix: nil, infix: binary, prec: PREC_TERM},
		TOKEN_AMPERSAND:     {prefix: nil, infix: binary, prec: PREC_TERM},
		TOKEN_SEMICOLON:     {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_EOL:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_SLASH:         {prefix: nil, infix: binary, prec: PREC_FACTOR},
		TOKEN_STAR:          {prefix: nil, infix: binary, prec: PREC_FACTOR},
		TOKEN_PERCENT:       {prefix: nil, infix: binary, prec: PREC_FACTOR},
		TOKEN_BANG:          {prefix: unary, infix: nil, prec: PREC_NONE},
		TOKEN_BANG_EQUAL:    {prefix: nil, infix: binary, prec: PREC_EQUALITY},
		TOKEN_EQUAL:         {prefix: nil, infix: nil, prec: PREC_ASSIGNMENT},
		TOKEN_EQUAL_EQUAL:   {prefix: nil, infix: binary, prec: PREC_EQUALITY},
		TOKEN_IN:            {prefix: nil, infix: binary, prec: PREC_EQUALITY},
		TOKEN_GREATER:       {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_GREATER_EQUAL: {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_LESS:          {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_LESS_EQUAL:    {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_IDENTIFIER:    {prefix: variable, infix: nil, prec: PREC_NONE},
		TOKEN_STRING:        {prefix: loxstring, infix: nil, prec: PREC_NONE},
		TOKEN_FLOAT:         {prefix: float, infix: nil, prec: PREC_NONE},
		TOKEN_INT:           {prefix: int_, infix: nil, prec: PREC_NONE},
		TOKEN_AND:           {prefix: nil, infix: and_, prec: PREC_AND},
		TOKEN_CLASS:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_ELSE:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FALSE:         {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_FOR:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FOREACH:       {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FUNC:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_IF:            {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_TRY:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_EXCEPT:        {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FINALLY:       {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_NIL:           {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_OR:            {prefix: nil, infix: or_, prec: PREC_OR},
		TOKEN_PRINT:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_RETURN:        {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_RAISE:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_SUPER:         {prefix: super, infix: nil, prec: PREC_NONE},
		TOKEN_THIS:          {prefix: this, infix: nil, prec: PREC_NONE},
		TOKEN_TRUE:          {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_VAR:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_CONST:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_WHILE:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_ERROR:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_EOF:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_STR:           {prefix: str_, infix: nil, prec: PREC_NONE},
	}
}

// match checks if the current token matches the specified token type.
// If it matches and is not EOF, it advances to the next token and returns true.
// This is the primary method for consuming expected tokens during parsing.
// TOKEN_SEMICOLON matches both actual semicolons and end-of-line tokens.
func (p *Parser) match(tt TokenType) bool {

	if !p.check(tt) {
		return false
	}
	if tt != TOKEN_EOF {
		p.advance()
	}
	return true
}

// check tests if the current token matches the specified token type without consuming it.
// This is used for lookahead during parsing to make decisions about which parsing path to take.
// Special case: TOKEN_SEMICOLON also matches TOKEN_EOL since both terminate statements.
func (p *Parser) check(tt TokenType) bool {

	return p.current.Tokentype == tt || (tt == TOKEN_SEMICOLON && p.current.Tokentype == TOKEN_EOL)
}

// checkNext peeks at the next token in the stream without consuming the current token.
// This provides limited lookahead capability for parsing decisions that require
// examining tokens beyond the current position.
func (p *Parser) checkNext(tt TokenType) bool {

	return p.scn.Tokens.At(p.scn.TokenIdx).Tokentype == tt
}

// checkAhead peeks at a token at the specified offset from the current position.
// This allows checking multiple tokens ahead for complex parsing decisions
// that require examining several upcoming tokens.
func (p *Parser) checkAhead(tt TokenType, offset int) bool {

	return p.scn.Tokens.At(p.scn.TokenIdx+offset).Tokentype == tt
}

// advance moves to the next token in the input stream, storing the current token as previous.
// It skips over any error tokens by reporting them and continuing to the next valid token.
// This ensures the parser always has a valid current token to work with.
func (p *Parser) advance() {

	p.previous = p.current
	for {
		p.current = p.scn.NextToken()
		if p.current.Tokentype != TOKEN_ERROR {
			break
		}
		p.errorAtCurrent(p.current.Lexeme())
	}

}

// getRule retrieves the parsing rule for a given token type from the rules table.
// Returns the ParseRule containing prefix/infix functions and precedence for the token.
// This is used by the Pratt parser to determine how to parse expressions.
func (p *Parser) getRule(tok TokenType) ParseRule {

	return p.rules[tok]
}

// declaration parses top-level declarations including classes, functions, variables, and constants.
// This is the main dispatch function for parsing global scope declarations.
// If an error occurs during parsing, it enters panic mode and synchronizes to recover.
func (p *Parser) declaration() {

	if p.match(TOKEN_CLASS) {
		p.classDeclaration()
	} else if p.match(TOKEN_FUNC) {
		p.funcDeclaration()
	} else if p.match(TOKEN_VAR) {
		p.varDeclaration(false)
	} else if p.match(TOKEN_CONST) {
		p.constDeclaration()
	} else {
		p.statement()
	}
	if p.panicMode {
		p.synchronize()
	}

}

// statement parses and compiles various statement types including control flow,
// exception handling, loops, conditionals, and expression statements.
// This is the main dispatch function for parsing executable statements within blocks.
func (p *Parser) statement() {

	if p.match(TOKEN_PRINT) {
		p.printStatement()
	} else if p.match(TOKEN_IMPORT) {
		p.importStatement()
	} else if p.match(TOKEN_FROM) {
		p.importFromStatement()
	} else if p.match(TOKEN_BREAK) {
		p.breakStatement()
	} else if p.match(TOKEN_BREAKPOINT) {
		p.breakpointStatement()
	} else if p.match(TOKEN_CONTINUE) {
		p.continueStatement()
	} else if p.match(TOKEN_TRY) {
		p.tryExceptStatement()
	} else if p.match(TOKEN_RAISE) {
		p.raiseStatement()
	} else if p.match(TOKEN_FOR) {
		p.forStatement()
	} else if p.match(TOKEN_FOREACH) {
		p.foreachStatement()
	} else if p.match(TOKEN_IF) {
		p.ifStatement()
	} else if p.match(TOKEN_RETURN) {
		p.returnStatement()
	} else if p.match(TOKEN_WHILE) {
		p.whileStatement()
	} else if p.match(TOKEN_LEFT_BRACE) {
		p.beginScope()
		p.block()
		p.endScope()
	} else {
		p.expressionStatement()
	}
}

// tryExceptStatement compiles try/except blocks for exception handling.
// Syntax: try { ... } except ExceptionType as var { ... } [except AnotherType as var2 { ... }]*
// It generates bytecode to:
// 1. Set up exception handling with OP_TRY
// 2. Execute the try block in a new scope
// 3. Handle multiple except clauses with different exception types
// 4. Bind caught exceptions to variables in except block scopes
// 5. Jump over except blocks if no exception occurs
func (p *Parser) tryExceptStatement() {

	_ = p.match(TOKEN_EOL)
	p.consume(TOKEN_LEFT_BRACE, "Expect '{' brace after try")
	exceptTry := p.emitTry()
	p.beginScope()
	p.block()
	p.endScope()
	endTryJump := p.emitJump(core.OP_END_TRY)
	for {
		p.consume(TOKEN_EXCEPT, "Expect except.")
		p.consume(TOKEN_IDENTIFIER, "Expect Exception type.")
		p.beginScope()
		idx := p.identifierConstant(p.previous)
		p.patchTry(exceptTry)
		p.consume(TOKEN_AS, "Expect as")
		ev := p.parseVariable("Expect exception variable name.")
		p.defineVariable(ev)
		_ = p.match(TOKEN_EOL)
		p.consume(TOKEN_LEFT_BRACE, "Expect left brace.")
		p.emitByte(core.OP_EXCEPT)
		p.emitByte(idx)
		p.block()
		p.endScope()
		p.emitByte(core.OP_END_EXCEPT)
		if !p.check(TOKEN_EXCEPT) {
			break
		}
	}

	p.patchJump(endTryJump)

}

// raiseStatement compiles raise statements for throwing exceptions.
// Syntax: raise expression;
// The expression should evaluate to an exception object that will be thrown.
// Generates OP_RAISE bytecode instruction to trigger exception handling.
func (p *Parser) raiseStatement() {

	p.expression() // this includes constructor calls
	p.consume(TOKEN_SEMICOLON, "Expect ';' after throw expression.")
	p.emitByte(core.OP_RAISE)
}

// importStatement parses import statements of the form:
// import module_name [as alias_name][, module_name [as alias_name]]*;
// It allows importing multiple modules in a single statement, separated by commas.
// Each module can optionally have an alias using the 'as' keyword.

// OP_IMPORT is used to import a module, takes two byte arguments
// - the first is the module name constant, the second is the alias constant.
func (p *Parser) importStatement() {
	c := 0
	for {
		p.consume(TOKEN_IDENTIFIER, "Expect module name.")
		nameConstant := p.identifierConstant(p.previous)
		c = c + 1
		p.emitBytes(core.OP_IMPORT, nameConstant)
		if p.match(TOKEN_AS) {
			p.consume(TOKEN_IDENTIFIER, "Expect alias name.")
			aliasConstant := p.identifierConstant(p.previous)
			p.emitByte(aliasConstant)
		} else {
			// no alias, use module name as alias
			p.emitByte(nameConstant)
		}
		if !p.match(TOKEN_COMMA) {
			break
		}
	}
	p.consume(TOKEN_SEMICOLON, "Expect ';' after import list.")
}

// importFromStatement parses import statements of the form:
// from module_name import name1, name2, ...
// from module_name import *
// It allows importing specific names from a module or all names using '*'.

func (p *Parser) importFromStatement() {

	p.consume(TOKEN_IDENTIFIER, "Expect module name.")
	nameConstant := p.identifierConstant(p.previous)
	p.emitBytes(core.OP_IMPORT_FROM, nameConstant)
	p.consume(TOKEN_IMPORT, "Expect 'import' after module name.")
	if p.match(TOKEN_STAR) {
		p.emitByte(0) // 0 means import all names
		p.consume(TOKEN_SEMICOLON, "Expect ';' after import list.")
		return
	}
	var names []Token
	for {
		p.consume(TOKEN_IDENTIFIER, "Expect name.")
		name := p.previous
		names = append(names, name)
		if !p.match(TOKEN_COMMA) {
			break
		}
	}
	// emit names
	length := len(names)
	p.emitByte(uint8(length)) // number of names to import
	for _, name := range names {
		constant := p.identifierConstant(name)
		p.emitByte(constant) // emit the constant for each name
	}
	p.consume(TOKEN_SEMICOLON, "Expect ';' after import list.")
}

// expression parses and compiles expressions starting with assignment precedence.
// This is the main entry point for parsing any expression in the language.
// Uses Pratt parsing to handle operator precedence correctly.
func (p *Parser) expression() {

	p.parsePrecedence(PREC_ASSIGNMENT)
}

// block parses and compiles a sequence of declarations/statements within braces.
// Continues parsing until reaching the closing brace or EOF.
// Expects the opening brace to already be consumed and consumes the closing brace.
// Allows optional end-of-line tokens after the closing brace.
func (p *Parser) block() {

	for !p.check(TOKEN_RIGHT_BRACE) && !p.check(TOKEN_EOF) {
		p.declaration()
	}
	p.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
	p.match(TOKEN_EOL) // allow EOL after block

}

// funcDeclaration parses and compiles function declarations.
// Creates a global variable for the function name and compiles the function body.
// The function is marked as initialized before compilation to allow recursive calls.
func (p *Parser) funcDeclaration() {

	global := p.parseVariable("Expect function name.")
	p.markInitialised()
	p.function(TYPE_FUNCTION)
	p.defineVariable(global)
}

// function compiles a function definition with the specified type (function, method, initializer).
// Creates a new compiler context for the function scope, parses parameters and body,
// then generates a closure with proper upvalue handling for captured variables.
// Function parameters are limited to 255 and become local variables in the function scope.
// function compiles function declarations and expressions.
// Creates a new compiler context for the function scope, parses parameters,
// compiles the function body, and generates a closure object with upvalue bindings.
// Handles parameter limits, nested scopes, and proper closure variable capture.
func (p *Parser) function(type_ FunctionType) {

	compiler := NewCompiler(type_, p.currentCompiler.scriptName, p.currentCompiler, p.currentCompiler.environment)
	p.currentCompiler = compiler
	funcname := p.previous.Lexeme()

	compiler.function.Name = core.MakeStringObject(funcname)

	p.beginScope()

	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after function name.")
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			p.currentCompiler.function.Arity += 1
			if p.currentCompiler.function.Arity > 255 {
				p.errorAtCurrent("Can't have more than 255 parameters")
			}
			constant := p.parseVariable("Expect parameter name.")
			p.defineVariable(constant)
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	p.match(TOKEN_EOL)
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after function parameters.")
	p.match(TOKEN_EOL) // allow EOL after parameters
	p.consume(TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	p.block()

	function := p.endCompiler()
	p.emitBytes(core.OP_CLOSURE, p.MakeConstant(core.MakeObjectValue(function, false)))

	for i := 0; i < function.UpvalueCount; i += 1 {
		uv := *(compiler.upvalues[i])
		if uv.isLocal {
			p.emitByte(1)
		} else {
			p.emitByte(0)
		}
		p.emitByte(uv.index)
	}
}

// classDeclaration parses and compiles class declarations with optional inheritance.
// Creates a class object, handles superclass inheritance, sets up class scope,
// compiles methods (including static methods), and manages the "super" keyword for inherited classes.
// Class names cannot inherit from themselves and superclasses cannot be from imported modules.
func (p *Parser) classDeclaration() {

	p.consume(TOKEN_IDENTIFIER, "Expect class name.")
	className := p.previous
	nameConstant := p.identifierConstant(p.previous)
	p.declareVariable()

	p.emitBytes(core.OP_CLASS, nameConstant)
	p.defineVariable(nameConstant)

	cc := &ClassCompiler{
		enclosing:     p.currentClass,
		hasSuperClass: false,
	}
	p.currentClass = cc

	if p.match(TOKEN_LESS) {
		p.consume(TOKEN_IDENTIFIER, "Expect superclass name.")
		variable(p, false)
		if p.identifiersEqual(className, p.previous) {
			p.error("A class cannot inherit from itself.")
		}
		if p.check(TOKEN_DOT) {
			p.error("Super class cannot be in an imported module (for now).")
		}
		p.beginScope()
		p.addLocal(SyntheticToken("super"))
		p.defineVariable(0)
		p.namedVariable(className, false)
		p.emitByte(core.OP_INHERIT)
		p.currentClass.hasSuperClass = true
	}

	p.namedVariable(className, false)
	p.match(TOKEN_EOL) // allow EOL after parameters
	p.consume(TOKEN_LEFT_BRACE, "Expect '{' before class body.")
	for !p.check(TOKEN_RIGHT_BRACE) && !p.check(TOKEN_EOF) {
		p.method()
	}
	p.consume(TOKEN_RIGHT_BRACE, "Expect '}' after class body.")
	p.match(TOKEN_EOL) // allow EOL after block
	p.emitByte(core.OP_POP)
	if p.currentClass.hasSuperClass {
		p.endScope()
	}
	p.currentClass = p.currentClass.enclosing
}

// method parses and compiles class methods including static methods and initializers.
// Static methods are bound to the class rather than instances.
// The "init" method is treated as a special initializer (constructor) that cannot be static.
// Regular methods are bound to class instances and have access to "this".
func (p *Parser) method() {

	static := false
	if p.match(TOKEN_STATIC) {
		static = true
	}

	p.consume(TOKEN_IDENTIFIER, "Expect method name.")
	constant := p.identifierConstant(p.previous)
	_type := TYPE_METHOD

	if p.previous.Lexeme() == "init" {
		if static {
			p.error("Static initialisers are not allowed.")
		}
		_type = TYPE_INITIALIZER
	}
	p.function(_type)
	if static {
		p.emitBytes(core.OP_STATIC_METHOD, constant)
		return
	}
	p.emitBytes(core.OP_METHOD, constant)
}

// varDeclaration parses and compiles variable declarations with optional initialization.
// Variables without explicit initialization are set to nil.
// The in_foreach parameter indicates if this is being used in a foreach loop
// where semicolons are not required after the variable declaration.
func (p *Parser) varDeclaration(in_foreach bool) {

	variable := p.parseVariable("Expect variable name")

	if p.match(TOKEN_EQUAL) {
		p.expression()
	} else {
		p.emitByte(core.OP_NIL) // empty local slot
	}
	if !in_foreach {
		p.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration")
	}

	p.defineVariable(variable)
}

// constDeclaration parses and compiles constant declarations.
// Constants must be initialized with a value and cannot be reassigned later.
// Creates an immutable variable binding that generates compile-time errors on reassignment attempts.
func (p *Parser) constDeclaration() {

	v := p.parseVariable("Expect variable name")

	if p.match(TOKEN_EQUAL) {
		p.expression()
	} else {
		p.error("Constants must be initialised.")
	}
	p.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration")

	p.defineConstVariable(v)
}

// isVariableDefined checks if a variable is already defined in the current scope or as a global.
// For local scopes, it checks local variables, upvalues, and globals.
// For global scope, it only checks the globals map.
// This is used to detect redefinition errors and support implicit variable declarations.
func (p *Parser) isVariableDefined(name Token, lexeme string) bool {

	var rv bool
	// core.LogFmt(core.DEBUG, "Checking if variable %s is defined\n", lexeme)
	// core.LogFmt(core.DEBUG, "p.globals: %v\n", p.globals)
	if p.currentCompiler.scopeDepth > 0 {
		rv = !(p.resolveLocal(p.currentCompiler, name) == -1 &&
			p.resolveUpvalue(p.currentCompiler, name) == -1 &&
			!p.checkGlobals(lexeme))
	} else {
		rv = p.checkGlobals(lexeme)
	}
	// core.LogFmt(core.DEBUG, "Variable %s is defined: %t\n", lexeme, rv)
	return rv
}

// handleImplicitDeclaration checks for and handles implicit variable declarations.
// In Lox, assignments like "a = 5" can create new variables if 'a' doesn't exist.
// For local scope: creates a new local variable declaration
// For global scope: adds to globals map and creates a variable declaration
// Returns true if an implicit declaration was handled, false otherwise.
func (p *Parser) handleImplicitDeclaration() bool {
	if p.check(TOKEN_IDENTIFIER) && p.checkNext(TOKEN_EQUAL) {
		name := p.current
		l := name.Lexeme()
		if !p.isVariableDefined(name, l) {
			if p.currentCompiler.scopeDepth > 0 {
				p.varDeclaration(false)
			} else {
				// core.LogFmt(core.DEBUG, "Implicitly declaring global variable %s\n", l)
				p.globals[l] = true
				p.varDeclaration(false)
			}
			return true
		}
	}
	return false
}

// handleUnpackingAssignment parses and compiles tuple/list unpacking assignments.
// Syntax: a, b, c = expr;
// This allows unpacking multiple values from a list or tuple into separate variables.
// It expects identifiers separated by commas on the left side of the assignment.
// For example: a, b, c = [1, 2, 3] or a, b, c = (4, 5, 6).
// Generates OP_UNPACK bytecode followed by the count of variables to unpack.
// The right-hand side expression is evaluated first, then OP_UNPACK distributes values.
func (p *Parser) handleUnpackingAssignment() bool {

	if p.check(TOKEN_IDENTIFIER) && p.checkNext(TOKEN_COMMA) {
		var names []Name
		// Parse identifiers separated by commas
		for {
			names = append(names, Name{p.current, p.current.Lexeme()})
			p.advance()
			if !p.match(TOKEN_COMMA) {
				break
			}
			if !p.check(TOKEN_IDENTIFIER) {
				p.errorAtCurrent("Expect variable name in unpacking assignment.")
				return false
			}
		}
		p.consume(TOKEN_EQUAL, "Expect '=' after unpacking variables.")
		p.expression() // Parse RHS expression

		// Emit unpacking assignment opcode
		p.emitByte(core.OP_UNPACK)
		p.emitByte(uint8(len(names)))

		// Assign to each variable.
		//  - locals we can update in place
		//  - globals we need to set with OP_SET_GLOBAL in reverse order
		if p.currentCompiler.scopeDepth > 0 {
			for _, name := range names {
				if !p.isVariableDefined(name.Token, name.Str) {
					if p.currentCompiler.scopeDepth > 0 {
						p.addLocal(name.Token)
						p.markInitialised()
					}
				}
			}
		} else {
			for i := len(names) - 1; i >= 0; i-- {
				name := names[i]
				if !p.isVariableDefined(name.Token, name.Str) {
					//core.LogFmt(core.DEBUG, "Implicitly declaring global variable %s\n", name.Str)
					p.globals[name.Str] = true
					arg := int(p.identifierConstant(name.Token))
					p.emitBytes(core.OP_SET_GLOBAL, uint8(arg))
					p.emitByte(core.OP_POP) // Pop the value off the stack after assignment
				}
			}
		}

		p.consume(TOKEN_SEMICOLON, "Expect ';' after unpacking assignment.")
		return true
	}
	return false
}

// expressionStatement parses and compiles expression statements.
// First tries to handle special cases (increment, unpacking, implicit declarations),
// then falls back to parsing a general expression followed by a semicolon.
// The expression result is popped from the stack since it's not used.
func (p *Parser) expressionStatement() {

	if p.handleUnpackingAssignment() {
		return
	}
	if p.handleImplicitDeclaration() {
		return
	}
	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
	p.emitByte(core.OP_POP)
}

// ifStatement compiles if-else conditional statements.
// Generates bytecode for condition evaluation, conditional jumps, and optional else clause.
// Uses jump patching to handle forward references to code locations not yet known.
// Stack management ensures the condition value is popped in both branches.
func (p *Parser) ifStatement() {

	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect '(' after condition.")

	thenJump := p.emitJump(core.OP_JUMP_IF_FALSE)
	p.emitByte(core.OP_POP)
	p.statement()
	elseJump := p.emitJump(core.OP_JUMP)
	p.patchJump(thenJump)
	p.emitByte(core.OP_POP)
	if p.match(TOKEN_ELSE) {
		p.statement()
	}
	p.patchJump(elseJump)

}

// returnStatement compiles return statements with optional return values.
// Validates that returns are only used within functions (not top-level script).
// Initializers cannot return explicit values (only implicit 'this').
// Empty returns are handled by emitReturn(), which returns appropriate default values.
func (p *Parser) returnStatement() {

	if p.currentCompiler.type_ == TYPE_SCRIPT {
		p.error("Can't return from top-level code.")
	}
	if p.match(TOKEN_SEMICOLON) {
		p.emitReturn()
	} else {
		if p.currentCompiler.type_ == TYPE_INITIALIZER {
			p.error("Can't return from an initializer.")
		}
		p.expression()
		p.consume(TOKEN_SEMICOLON, "Expect ';' after return value.")
		op := core.OP_RETURN

		p.emitByte(op)
	}
}

// whileStatement compiles while loops with break/continue support.
// Sets up a new loop context to track break statements and continue jumps.
// Generates loop bytecode with proper jump handling for condition evaluation and loop body.
// Restores the previous loop context when compilation completes.
func (p *Parser) whileStatement() {

	loopSave := p.currentCompiler.loop
	p.currentCompiler.loop = NewLoop()

	p.currentCompiler.loop.start = len(p.currentChunk().Code)
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after while.")
	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := p.emitJump(core.OP_JUMP_IF_FALSE)
	p.emitByte(core.OP_POP)
	p.statement()
	p.emitLoop(core.OP_LOOP, p.currentCompiler.loop.start)
	if len(p.currentCompiler.loop.breaks) != 0 {
		for _, jump := range p.currentCompiler.loop.breaks {
			p.patchJump(jump)
		}
	}
	p.patchJump(exitJump)
	p.emitByte(core.OP_POP)

	p.currentCompiler.loop = loopSave
}

// forStatement compiles traditional for loops with initialization, condition, and increment.
// Syntax: for (init; condition; increment) statement
// Creates a new scope for loop variables and manages loop control flow.
// Handles optional clauses and generates proper bytecode for loop execution and jump management.
func (p *Parser) forStatement() {

	loopSave := p.currentCompiler.loop
	p.currentCompiler.loop = NewLoop()

	p.beginScope()
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after for.")
	// initialiser
	if p.match(TOKEN_SEMICOLON) {
	} else if p.match(TOKEN_VAR) {
		p.varDeclaration(false)
	} else {
		p.expressionStatement()
	}
	p.currentCompiler.loop.start = len(p.currentChunk().Code)
	// exit condition
	exitJump := -1
	if !p.match(TOKEN_SEMICOLON) {
		p.expression()
		p.consume(TOKEN_SEMICOLON, "Expect ';'.")
		exitJump = p.emitJump(core.OP_JUMP_IF_FALSE)
		p.emitByte(core.OP_POP)
	}
	// increment
	if !p.match(TOKEN_RIGHT_PAREN) {
		// jump over increment, will be executed after body
		bodyJump := p.emitJump(core.OP_JUMP)
		incrementStart := len(p.currentChunk().Code)
		p.expression()
		p.emitByte(core.OP_POP)
		p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")
		p.emitLoop(core.OP_LOOP, p.currentCompiler.loop.start)
		p.currentCompiler.loop.start = incrementStart
		p.patchJump(bodyJump)
	}
	p.match(TOKEN_EOL)
	p.statement()
	if len(p.currentCompiler.loop.breaks) != 0 {
		for _, jump := range p.currentCompiler.loop.breaks {
			p.patchJump(jump)
		}
	}
	p.emitLoop(core.OP_LOOP, p.currentCompiler.loop.start)

	if exitJump != -1 {
		p.patchJump(exitJump)
		p.emitByte(core.OP_POP)
	}
	p.endScope()
	p.currentCompiler.loop = loopSave
}

// breakStatement compiles break statements for exiting loops early.
// Validates that break is only used within loops.
// Cleans up local variables on the stack before jumping out of the loop.
// Records the jump location for later patching when the loop end is known.
func (p *Parser) breakStatement() {

	p.consume(TOKEN_SEMICOLON, "Expect ';' after statement.")
	if p.currentCompiler.loop == nil {
		p.errorAtCurrent("Cannot use break outside loop.")
	}

	// drop local vars on stack
	c := p.currentCompiler

	for i := 0; i < c.localCount; i += 1 {
		if c.locals[i].depth >= c.scopeDepth-1 {
			p.emitByte(core.OP_POP)
		}
	}
	p.currentCompiler.loop.breaks = append(p.currentCompiler.loop.breaks, p.emitJump(core.OP_JUMP))
}

// breakpointStatement compiles breakpoint statements for debugging support.
// Generates OP_BREAKPOINT bytecode instruction that can be used by debuggers
// to pause execution at specific points in the code for inspection.
func (p *Parser) breakpointStatement() {

	p.consume(TOKEN_SEMICOLON, "Expect ';' after statement.")
	p.emitByte(core.OP_BREAKPOINT)
}

// continueStatement compiles continue statements for skipping to the next loop iteration.
// Validates that continue is only used within loops.
// Cleans up local variables on the stack before jumping.
// For foreach loops, uses a forward jump; for regular loops, jumps back to loop start.
func (p *Parser) continueStatement() {

	p.consume(TOKEN_SEMICOLON, "Expect ';' after statement.")
	if p.currentCompiler.loop == nil {
		p.errorAtCurrent("Cannot use continue outside loop.")
	}

	// drop local vars on stack
	c := p.currentCompiler
	for i := 0; i < c.localCount; i += 1 {
		if c.locals[i].depth >= c.scopeDepth-1 {
			p.emitByte(core.OP_POP)
		}
	}
	if p.currentCompiler.loop.foreach {
		p.currentCompiler.loop.continue_ = p.emitJump(core.OP_JUMP)
	} else {
		p.emitLoop(core.OP_LOOP, p.currentCompiler.loop.start)
	}

}

// foreachStatement compiles foreach loops for iterating over collections.
// Syntax: foreach (var item in collection) statement
// Creates 3 local variables on the stack:
//   - var receiving iterator output (user-visible)
//   - iterated list/string (hidden)
//   - iteration index (hidden)
//
// Uses OP_FOREACH bytecode for efficient iteration over lists, strings, and iterables.
func (p *Parser) foreachStatement() {

	loopSave := p.currentCompiler.loop
	p.currentCompiler.loop = NewLoop()
	p.currentCompiler.loop.foreach = true // so continue knows to jump to next

	p.beginScope()
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after for.")
	p.match(TOKEN_VAR)
	p.varDeclaration(true)
	slot := p.currentCompiler.localCount - 1
	p.consume(TOKEN_IN, "Expect in after foreach variable.")

	// get iterator and put it in a temp local
	p.expression()
	p.addLocal(SyntheticToken("__iter"))
	iterSlot := p.currentCompiler.localCount - 1
	p.markInitialised()

	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after iterable.")

	jumpToEnd := p.emitForeach(uint8(slot), uint8(iterSlot))
	// each iteration will jump back to this point
	p.currentCompiler.loop.start = len(p.currentChunk().Code)
	// body of foreach
	p.statement()
	// if it contained a continue, patch its jump to come here
	if p.currentCompiler.loop.continue_ != 0 {
		p.patchJump(p.currentCompiler.loop.continue_)
	}
	// jump to loop start
	p.emitLoop(core.OP_NEXT, p.currentCompiler.loop.start)
	p.emitByte(uint8(iterSlot))
	// iteration complete, patch foreach to come here
	p.emitByte(core.OP_END_FOREACH)
	p.patchForeach(jumpToEnd)
	// did the body contain a break? if so patch its jump to come here
	if len(p.currentCompiler.loop.breaks) != 0 {
		for _, jump := range p.currentCompiler.loop.breaks {
			p.patchJump(jump)
		}
	}
	p.endScope()
	p.currentCompiler.loop = loopSave
}

// printStatement compiles print statements for outputting values.
// Converts the expression result to a string and prints it.
// Uses OP_STR to ensure proper string conversion followed by OP_PRINT.
func (p *Parser) printStatement() {

	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	p.emitByte(core.OP_STR)
	p.emitByte(core.OP_PRINT)
}

// synchronize performs error recovery by advancing tokens until a safe synchronization point.
// Called after parse errors to resynchronize the parser at statement boundaries.
// Stops at statement-starting keywords or after semicolons/end-of-lines.
// This allows the parser to continue and report multiple errors in one pass.
func (p *Parser) synchronize() {

	p.panicMode = false
	for p.current.Tokentype != TOKEN_EOF {
		if p.previous.Tokentype == TOKEN_SEMICOLON || p.previous.Tokentype == TOKEN_EOL {
			return
		}
		switch p.current.Tokentype {
		case TOKEN_CLASS:
			return
		case TOKEN_FUNC:
			return
		case TOKEN_FOR:
			return
		case TOKEN_VAR:
			return
		case TOKEN_IF:
			return
		case TOKEN_WHILE:
			return
		case TOKEN_PRINT:
			return
		case TOKEN_RETURN:
			return
		}
		p.advance()
	}
}

// consume advances to the next token if it matches the expected type, or reports an error.
// This is the primary mechanism for enforcing syntax requirements during parsing.
// Special handling for TOKEN_SEMICOLON allows TOKEN_EOL as an alternative (optional semicolons).
func (p *Parser) consume(toktype TokenType, msg string) {

	if p.current.Tokentype == toktype || (toktype == TOKEN_SEMICOLON && p.current.Tokentype == TOKEN_EOL) {
		p.advance()
		return
	}
	p.errorAtCurrent(msg)
}

// emitByte writes a single bytecode instruction to the current chunk.
// Records the line number from the previous token for debugging information.
// This is the fundamental method for generating bytecode during compilation.
func (p *Parser) emitByte(byte uint8) {

	p.currentChunk().WriteOpCode(byte, p.previous.Line)
}

// emitBytes writes two consecutive bytecode instructions to the current chunk.
// Convenience function for operations that require an opcode followed by an operand.
func (p *Parser) emitBytes(byte1, byte2 uint8) {

	p.emitByte(byte1)
	p.emitByte(byte2)
}

// emitLoop generates a backward jump instruction for loops.
// Calculates the offset from the current position back to the loop start.
// Emits the loop instruction followed by a 16-bit offset for the jump distance.
// Reports an error if the loop body exceeds the maximum jump distance.
func (p *Parser) emitLoop(instr uint8, loopStart int) {

	p.emitByte(instr)

	offset := len(p.currentChunk().Code) - loopStart + 2
	if offset >= int(^uint16(0)) {
		p.error("Loop body too large")
	}

	p.emitByte(uint8((offset >> 8) & 0xff))
	p.emitByte(uint8(offset & 0xff))
}

// emitJump generates a forward jump instruction with placeholder offset.
// Emits the jump instruction followed by 0xffff as a placeholder for the jump distance.
// Returns the offset where the jump target needs to be patched later.
// Used for conditional jumps and control flow where the target isn't known yet.
func (p *Parser) emitJump(instr uint8) int {

	p.emitByte(instr)
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.currentChunk().Code) - 2
}

// emitForeach generates bytecode for foreach loop initialization.
// Emits OP_FOREACH followed by variable slot, iterator slot, and placeholder jump offset.
// Returns the offset for later patching when the foreach loop end is known.
// The instruction sets up iteration state and prepares for loop execution.
func (p *Parser) emitForeach(slot uint8, iterslot uint8) int {

	p.emitByte(core.OP_FOREACH)
	p.emitByte(slot)
	p.emitByte(iterslot)
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.currentChunk().Code) - 3
}

// emitTry generates bytecode for try block initialization in exception handling.
// Emits OP_TRY followed by a placeholder jump offset to be patched later.
// Returns the offset for patching when the corresponding except block location is known.
// Sets up exception handling context for the try block.
func (p *Parser) emitTry() int {

	p.emitByte(core.OP_TRY)
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.currentChunk().Code) - 2
}

// currentChunk returns the bytecode chunk being compiled for the current function.
// This provides access to the chunk where bytecode instructions and constants are stored.
func (p *Parser) currentChunk() *core.Chunk {

	return p.currentCompiler.function.Chunk
}

// endCompiler finalizes compilation of the current function and returns to the enclosing compiler.
// Emits a return instruction, and runs peep-hole optimizations on the current chunk.
// optionally disassembles the generated bytecode for debugging,
// and restores the previous compiler context. Returns the completed function object.
func (p *Parser) endCompiler() *core.FunctionObject {

	p.emitReturn()

	p.peepHoleOptimise()

	function := p.currentCompiler.function
	s := ""
	if function.Name.Get() == "" {
		s = p.currentCompiler.scriptName
	} else {
		s = function.Name.Get()
	}
	if core.DebugPrintCode && !core.DebugSuppress {
		if !p.hadError {
			debug.Disassemble(p.currentChunk(), s)
		}
	}

	p.currentCompiler = p.currentCompiler.enclosing
	return function
}

// beginScope enters a new lexical scope by incrementing the scope depth.
// Local variables declared in this scope will have this depth level.
// Used for blocks, functions, and other constructs that create new scopes.
func (p *Parser) beginScope() {

	p.currentCompiler.scopeDepth += 1
}

// endScope exits the current lexical scope and cleans up local variables.
// Decrements scope depth and removes local variables that belong to this scope.
// Emits OP_CLOSE_UPVALUE for captured variables or OP_POP for regular locals.
// This ensures proper stack cleanup and upvalue closure when exiting scopes.
func (p *Parser) endScope() {

	c := p.currentCompiler
	c.scopeDepth--

	// drop local vars on stack
	for c.localCount > 0 && c.locals[c.localCount-1].depth > c.scopeDepth {
		if c.locals[c.localCount-1].isCaptured {
			p.emitByte(core.OP_CLOSE_UPVALUE)
		} else {
			p.emitByte(core.OP_POP)
		}
		// update the end IP of the local variable
		for i := len(p.currentChunk().LocalVars) - 1; i >= 0; i-- {
			lv := &p.currentChunk().LocalVars[i]
			if lv.Slot == c.localCount-1 && lv.EndIp == -1 {
				lv.EndIp = len(p.currentChunk().Code)
				break
			}
		}
		c.localCount--
	}
}

// parsePrecedence implements Pratt parsing for expressions with operator precedence.
// Starts with a prefix expression, then processes infix operators based on precedence.
// The precedence parameter controls how tightly the current expression binds.
// Handles assignment validation and ensures proper left-to-right associativity.
func (p *Parser) parsePrecedence(prec Precedence) {

	p.advance()

	prefixRule := p.getRule(p.previous.Tokentype).prefix
	if prefixRule == nil {
		p.error("Expect expression")
		return
	}

	canAssign := prec <= PREC_ASSIGNMENT
	prefixRule(p, canAssign)
	for prec <= p.getRule(p.current.Tokentype).prec {

		p.advance()
		infixRule := p.getRule(p.previous.Tokentype).infix
		if infixRule != nil {

			infixRule(p, canAssign)
		}

	}
	// if = is left over, no rule it, return an error.
	if canAssign && p.match(TOKEN_EQUAL) {
		p.error("Invalid assignment target.")
	}
}

// identifierConstant creates a string constant from a token and adds it to the constant pool.
// Converts the token's lexeme to a string object value and returns its constant index.
// Used for variable names, method names, and other identifiers that need runtime lookup.
func (p *Parser) identifierConstant(t Token) uint8 {

	s := t.Lexeme()
	v := core.MakeStringObjectValue(s, false)
	return p.MakeConstant(v)
}

// identifiersEqual compares two tokens for identifier equality.
// Checks both length and lexeme content for efficient string comparison.
// Used for variable resolution and scope management during compilation.
func (p *Parser) identifiersEqual(a, b Token) bool {

	if a.Length != b.Length {
		return false
	}
	if a.Lexeme() != b.Lexeme() {
		return false
	}
	return true
}

// resolveLocal searches for a local variable by name in the current function's scope.
// Searches backwards through local variables to implement proper shadowing semantics.
// Returns the slot index if found, or -1 if not found.
// Prevents reading variables in their own initializer to catch use-before-definition errors.
func (p *Parser) resolveLocal(compiler *Compiler, name Token) int {

	for i := compiler.localCount - 1; i >= 0; i-- {
		local := compiler.locals[i]
		if p.identifiersEqual(name, local.name) {
			if local.depth == -1 {
				p.error("Can't read local variable in its own initialiser.")
			}
			return i
		}
	}
	return -1
}

// addUpvalue adds an upvalue to a compiler's upvalue list for closure variable capture.
// Checks if the upvalue already exists to avoid duplicates.
// Returns the index of the upvalue in the function's upvalue array.
// Used to capture local variables and upvalues from enclosing scopes in closures.
func (p *Parser) addUpvalue(compiler *Compiler, index uint8, isLocal bool) int {

	upvalueCount := compiler.function.UpvalueCount

	// does upvalue already exist ?
	for i := 0; i < upvalueCount; i += 1 {
		upvalue := *(compiler.upvalues[i])
		if upvalue.index == index && upvalue.isLocal == isLocal {
			return i
		}
	}
	if upvalueCount == 256 {
		p.error("Too many closure variables in function.")
		return 0
	}
	uv := &Upvalue{
		isLocal: isLocal,
		index:   index,
	}
	compiler.upvalues[upvalueCount] = uv
	compiler.function.UpvalueCount += 1

	return upvalueCount

}

// resolveUpvalue recursively resolves variables from enclosing scopes for closure capture.
// Implements the upvalue resolution algorithm for lexical scoping and closure variable access.
// 1. Looks for local variable in immediately enclosing function (base case)
// 2. If not found, recursively searches outer scopes via resolveUpvalue calls
// 3. When found, creates upvalue chain back down to innermost function
// 4. Marks captured locals and distinguishes between local vs upvalue captures
// This enables proper closure semantics across function boundaries.
func (p *Parser) resolveUpvalue(compiler *Compiler, name Token) int {

	if compiler.enclosing == nil {
		return -1
	}
	local := p.resolveLocal(compiler.enclosing, name)
	if local != -1 {
		compiler.enclosing.locals[local].isCaptured = true
		return p.addUpvalue(compiler, uint8(local), true)
	}

	upValue := p.resolveUpvalue(compiler.enclosing, name)
	if upValue != -1 {
		return p.addUpvalue(compiler, uint8(upValue), false)
	}

	return -1
}

// parseVariable consumes an identifier token for variable declaration and handles scoping.
// Returns the constant table index for global variables, or 0 for local variables.
// Declares the variable in the current scope and validates the identifier token.
func (p *Parser) parseVariable(errorMsg string) uint8 {

	p.consume(TOKEN_IDENTIFIER, errorMsg)
	p.declareVariable()
	// if local, don't add to constant table
	if p.currentCompiler.scopeDepth > 0 {
		return 0
	}
	return p.identifierConstant(p.previous)
}

// markInitialised marks the most recently declared local variable as initialized.
// Sets the variable's depth to the current scope depth, making it accessible.
// Only applies to local variables; global variables are handled differently.
func (p *Parser) markInitialised() {

	c := p.currentCompiler
	if c.scopeDepth == 0 {
		return
	}
	c.locals[c.localCount-1].depth = c.scopeDepth
}

// setLocalImmutable marks the most recently added constant as immutable.
// Applies the immutable wrapper to prevent modification of const variables.
// Used for const declarations to enforce compile-time immutability.
func (p *Parser) setLocalImmutable() {

	c := p.currentChunk()
	c.Constants[len(c.Constants)-1] = core.Immutable(c.Constants[len(c.Constants)-1])
}

// defineVariable finalizes variable definition with appropriate bytecode emission.
// For local variables: marks as initialized (already on stack)
// For global variables: registers in globals map and emits OP_DEFINE_GLOBAL
// Handles the scope-dependent storage of variable definitions.
func (p *Parser) defineVariable(global uint8) {

	// if local, it will already be on the stack
	if p.currentCompiler.scopeDepth > 0 {
		p.markInitialised()
		return
	}
	x := p.currentChunk().Constants[global].AsString().Get()
	// core.LogFmt(core.DEBUG, "Adding global identifier '%s'\n", x)
	p.globals[x] = true
	p.emitBytes(core.OP_DEFINE_GLOBAL, global)
}

// argumentList parses function call arguments and returns the argument count.
// Handles comma-separated expression list within parentheses.
// Enforces the 255 argument limit and validates proper parentheses syntax.
// Returns the number of arguments parsed for the function call bytecode.
func (p *Parser) argumentList() uint8 {

	var argCount uint8 = 0
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			p.expression()
			argCount += 1
			if argCount == 255 {
				p.error("Can't have more than 255 arguments. ")
			}
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	p.match(TOKEN_EOL) // allow EOL after arguments
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after arguments")
	return argCount
}

// parseList parses list literal syntax and returns the element count.
// Handles comma-separated expressions within square brackets.
// Enforces the 255 element limit for list initialization.
// Returns the number of elements for the list creation bytecode.
func (p *Parser) parseList() uint8 {

	var itemCount uint8 = 0
	if !p.check(TOKEN_RIGHT_BRACKET) {
		for {
			p.expression()
			itemCount += 1
			if itemCount == 255 {
				p.error("Can't have more than 255 initialiser items. ")
			}
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	p.match(TOKEN_EOL) // allow EOL after list items
	p.consume(TOKEN_RIGHT_BRACKET, "Expect ']' after list items.")
	return itemCount
}

// parseDict parses dictionary literal syntax and returns the key-value pair count.
// Handles key:value pairs separated by commas within curly braces.
// Enforces the 255 key limit for dictionary initialization.
// Allows optional end-of-line tokens after dictionary items.
func (p *Parser) parseDict() uint8 {

	var itemCount uint8 = 0
	if !p.match(TOKEN_RIGHT_BRACE) {
		for {
			p.expression()
			p.consume(TOKEN_COLON, "Expect ':' after key.")
			p.expression()
			itemCount += 1
			if itemCount == 255 {
				p.error("Can't have more than 255 initialiser keys. ")
			}
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
		p.match(TOKEN_EOL) // allow EOL after dict items
		p.consume(TOKEN_RIGHT_BRACE, "Expect '}' after dictionary items.")
	}

	return itemCount
}

// defineConstVariable finalizes const variable definition with immutability.
// For local variables: marks as initialized and sets immutable flag
// For global variables: emits OP_DEFINE_GLOBAL_CONST for const semantics
// Ensures constant variables cannot be reassigned after definition.
func (p *Parser) defineConstVariable(global uint8) {

	// if local, it will already be on the stack
	if p.currentCompiler.scopeDepth > 0 {
		p.markInitialised()
		p.setLocalImmutable()
		return
	}
	p.emitBytes(core.OP_DEFINE_GLOBAL_CONST, global)
}

// declareVariable declares a new variable in the current scope.
// For global scope: no action needed (handled at definition time)
// For local scope: validates no duplicate names and adds to local array
// Prevents variable shadowing within the same scope level.
func (p *Parser) declareVariable() {

	if p.currentCompiler.scopeDepth == 0 {
		return
	}
	name := p.previous
	// check we are not trying to create 2 locals with same name
	// current scope is at end of array, check back from that
	for i := p.currentCompiler.localCount - 1; i >= 0; i-- {

		local := p.currentCompiler.locals[i]
		if local.depth != -1 && local.depth < p.currentCompiler.scopeDepth {
			break
		}
		if p.identifiersEqual(name, local.name) {
			p.error("Already a variable with this name in this scope.")
		}
	}
	p.addLocal(name)
}

// checkGlobals verifies if a variable name exists in the global scope.
// Used for validating global variable references and preventing undefined access.
// Returns true if the variable has been declared globally, false otherwise.
func (p *Parser) checkGlobals(name string) bool {
	_, ok := p.globals[name]
	return ok
}

// resolveVariable determines the scope and access method for a named variable.
// Returns the variable's index/slot and the appropriate get/set opcodes.
// Checks local scope first, then upvalues, finally global scope.
// Used by namedVariable to generate the correct variable access bytecode.
func (p *Parser) resolveVariable(name Token) (int, uint8, uint8) {

	// core.LogFmt(core.DEBUG, "namedVariable %s canAssign %t\n", name.Lexeme(), canAssign)
	var getOp, setOp uint8
	a := name.Lexeme()
	_ = a
	arg := p.resolveLocal(p.currentCompiler, name)
	if arg != -1 {
		getOp = core.OP_GET_LOCAL
		setOp = core.OP_SET_LOCAL
		// core.LogFmt(core.DEBUG, "Local variable %s found at index %d\n", name.Lexeme(), arg)
	} else if arg = p.resolveUpvalue(p.currentCompiler, name); arg != -1 {
		getOp = core.OP_GET_UPVALUE
		setOp = core.OP_SET_UPVALUE
		// core.LogFmt(core.DEBUG, "Upvalue %s found at index %d\n", name.Lexeme(), arg)

	} else {
		arg = int(p.identifierConstant(name))
		getOp = core.OP_GET_GLOBAL
		setOp = core.OP_SET_GLOBAL
		// core.LogFmt(core.DEBUG, "Global variable %s found at index %d\n", name.Lexeme(), arg)
	}
	return arg, getOp, setOp
}

// namedVariable handles variable access and assignment for a specific named variable.
// Resolves the variable scope (local, upvalue, or global) and emits appropriate bytecode.
// For assignment: parses the right-hand expression and emits set operation
// For access: emits get operation to load the variable value
func (p *Parser) namedVariable(name Token, canAssign bool) {

	arg, getOp, setOp := p.resolveVariable(name)

	if p.handleCompoundAssignment(canAssign, getOp, setOp, arg) {
		return
	}
	if canAssign && p.match(TOKEN_EQUAL) {
		p.expression()
		p.emitBytes(setOp, uint8(arg))
	} else {
		p.emitBytes(getOp, uint8(arg))
	}
}

func (p *Parser) handleCompoundAssignment(canAssign bool, getOp uint8, setOp uint8, arg int) bool {

	if canAssign && (p.check(TOKEN_PLUS_EQUAL) || p.check(TOKEN_MINUS_EQUAL)) {
		// Handle compound assignment
		opType := p.current.Tokentype
		p.advance() // consume += or -=

		// Get current value
		p.emitBytes(getOp, uint8(arg))

		// Parse right-hand side
		p.expression()

		// Perform the operation
		switch opType {
		case TOKEN_PLUS_EQUAL:
			p.emitByte(core.OP_ADD_NUMERIC)
		case TOKEN_MINUS_EQUAL:
			p.emitByte(core.OP_SUBTRACT)
		}

		// Store the result back
		p.emitBytes(setOp, uint8(arg))
		return true
	}
	return false
}

// addLocal adds a new local variable to the current function's local variable array.
// Enforces the 256 local variable limit and initializes the local with uninitialized state.
// Records variable information for debugging and scope management.
func (p *Parser) addLocal(name Token) {

	if p.currentCompiler.localCount == 256 {
		p.error("Too many variables in function")
		return
	}
	local := &Local{
		name:       name,
		lexeme:     name.Lexeme(),
		depth:      -1, // marks as uninitialised
		isCaptured: false,
	}
	p.currentCompiler.locals[p.currentCompiler.localCount] = local
	p.currentCompiler.localCount += 1
	// core.LogFmt(core.DEBUG, "Added local %d %s at depth %d\n", p.currentCompiler.localCount, local.lexeme, p.currentCompiler.scopeDepth)

	// add local var info to current chunk for debugging
	chunk := p.currentChunk()
	chunk.LocalVars = append(chunk.LocalVars, core.LocalVarInfo{
		Name:    name.Lexeme(),
		StartIp: len(chunk.Code),
		EndIp:   -1,
		Slot:    p.currentCompiler.localCount - 1,
	})

}

// emitConstant creates a constant value and emits bytecode to load it onto the stack.
// Adds the value to the constant table and generates OP_CONSTANT instruction.
// Used for literal values like numbers, strings, and other compile-time constants.
func (p *Parser) emitConstant(value core.Value) {

	p.emitBytes(core.OP_CONSTANT, p.MakeConstant(value))
}

// patchJump fills in the jump offset for a previously emitted jump instruction.
// Calculates the distance from the jump instruction to the current code position.
// Updates the placeholder bytes with the actual 16-bit jump distance.
// Reports an error if the jump distance exceeds the maximum 16-bit value.
func (p *Parser) patchJump(offset int) {

	jump := len(p.currentChunk().Code) - offset - 2
	if uint16(jump) > ^uint16(0) {
		p.error("Jump overflow")
	}
	p.currentChunk().Code[offset] = uint8((jump >> 8) & 0xff)
	p.currentChunk().Code[offset+1] = uint8(jump & 0xff)

}

// patchForeach fills in the jump offset for a foreach loop instruction.
// Similar to patchJump but handles the specific offset layout for foreach instructions.
// Updates bytes at offset+1 and offset+2 (skipping the foreach opcode byte).
// Used to patch the exit jump when foreach iteration completes.
func (p *Parser) patchForeach(offset int) {

	jump := len(p.currentChunk().Code) - offset - 2
	if uint16(jump) > ^uint16(0) {
		p.error("Jump overflow")
	}
	p.currentChunk().Code[offset+1] = uint8((jump >> 8) & 0xff)
	p.currentChunk().Code[offset+2] = uint8(jump & 0xff)

}

// patchTry patches a try instruction's jump offset to point to the except handler.
// Updates the placeholder bytes in the try instruction with the actual address
// where the except clause begins. Used in exception handling compilation.
func (p *Parser) patchTry(offset int) {

	address := len(p.currentChunk().Code)
	p.currentChunk().Code[offset] = uint8((address >> 8) & 0xff)
	p.currentChunk().Code[offset+1] = uint8(address & 0xff)
}

// MakeConstant adds a value to the constant table and returns its index.
// Used for literals, identifiers, and other constant values that need runtime access.
// Enforces the 255 constant limit per chunk and reports errors if exceeded.
func (p *Parser) MakeConstant(value core.Value) uint8 {

	constidx := p.currentChunk().AddConstant(value)
	if constidx > 254 {
		p.error("Too many constants in one chunk")
		return 0
	}
	return constidx
}

// emitReturn generates appropriate return bytecode based on function type.
// For initializers: returns 'this' (slot 0) automatically
// For other functions: returns nil by default
// Follows Lox semantics where initializers always return the instance.
func (p *Parser) emitReturn() {

	if p.currentCompiler.type_ == TYPE_INITIALIZER {
		p.emitBytes(core.OP_GET_LOCAL, 0)
	} else {
		p.emitByte(core.OP_NIL)
	}
	op := core.OP_RETURN

	p.emitByte(op)
}

// slice1 handles slice expressions starting with colon: a[:] or a[:exp]
// Implements Python-style slicing from beginning of sequence.
// Supports both read access (OP_SLICE) and assignment (OP_SLICE_ASSIGN).
// Uses nil as the start index to indicate slicing from the beginning.
func (p *Parser) slice1(canAssign bool) {
	// slice from -> stack
	p.emitByte(core.OP_NIL)
	if p.match(TOKEN_RIGHT_BRACKET) {
		// [:]
		if canAssign && p.match(TOKEN_EQUAL) {
			// slice to -> stack
			p.emitByte(core.OP_NIL)
			// RHS -> stack
			p.expression()
			p.emitByte(core.OP_SLICE_ASSIGN)
		} else {
			p.emitByte(core.OP_NIL)
			p.emitByte(core.OP_SLICE)
		}
	} else {
		// [:exp]
		// slice to -> stack
		p.expression()
		p.consume(TOKEN_RIGHT_BRACKET, "Expect ']' after expression.")
		if canAssign && p.match(TOKEN_EQUAL) {
			// RHS -> stack
			p.expression()
			p.emitByte(core.OP_SLICE_ASSIGN)
		} else {
			p.emitByte(core.OP_SLICE)
		}
	}
}

// index handles single-element indexing: a[exp]
// Supports both read access (OP_INDEX) and assignment (OP_INDEX_ASSIGN).
// Used for accessing individual elements in lists, strings, and dictionaries.
func (p *Parser) index(canAssign bool) {

	if canAssign && p.match(TOKEN_EQUAL) {
		// RHS -> stack
		p.expression()
		p.emitByte(core.OP_INDEX_ASSIGN)
	} else {
		p.emitByte(core.OP_INDEX)
	}
}

// slice2 handles slice expressions ending with colon: a[exp:] or a[exp:exp]
// Implements Python-style slicing from a start index to end or specified endpoint.
// Supports both read access (OP_SLICE) and assignment (OP_SLICE_ASSIGN).
// Uses nil as the end index when slicing to the end of the sequence.
func (p *Parser) slice2(canAssign bool) {

	if p.match(TOKEN_RIGHT_BRACKET) {
		// [exp:]
		p.emitByte(core.OP_NIL)
		if canAssign && p.match(TOKEN_EQUAL) {
			// RHS -> stack
			p.expression()
			p.emitByte(core.OP_SLICE_ASSIGN)
		} else {
			p.emitByte(core.OP_SLICE)
		}
	} else {
		// [exp:exp]
		p.expression()
		p.consume(TOKEN_RIGHT_BRACKET, "Expect ']' after expression")
		if canAssign && p.match(TOKEN_EQUAL) {
			// RHS -> stack
			p.expression()
			p.emitByte(core.OP_SLICE_ASSIGN)
		} else {
			p.emitByte(core.OP_SLICE)
		}
	}
}

// errorAtCurrent reports a compilation error at the current token position.
// Convenience wrapper around errorAt for errors at the current parsing position.
// Used when detecting syntax errors in the current token being processed.
func (p *Parser) errorAtCurrent(msg string) {

	p.errorAt(p.current, msg)
}

// error reports a compilation error at the previous token position.
// Convenience wrapper around errorAt for errors related to the last consumed token.
// Most commonly used for syntax and semantic errors during parsing.
func (p *Parser) error(msg string) {

	p.errorAt(p.previous, msg)
}

// errorAt reports a compilation error at a specific token location.
// Formats and prints error messages with file name, line number, and context.
// Activates panic mode to prevent cascading errors during error recovery.
// Includes special handling for EOF and ERROR tokens for better diagnostics.
func (p *Parser) errorAt(tok Token, msg string) {

	if p.panicMode {
		return
	}
	p.panicMode = true
	fmt.Printf("In %s: ", p.currentCompiler.scriptName)
	fmt.Printf("[line %d] Error ", tok.Line)
	if tok.Tokentype == TOKEN_EOF {
		fmt.Printf(" at end")
	} else if tok.Tokentype == TOKEN_ERROR {
		fmt.Printf(" at %s ", tok.Lexeme())
	} else {
		fmt.Printf(" at %s ", tok.Lexeme())
	}
	fmt.Printf(" : %s\n", msg)
	p.hadError = true
}

//=============================================================================
// pratt parser functions

// binary handles all binary infix operators with proper precedence.
// Parses the right operand with appropriate precedence (left-associative: prec + 1).
// Emits the corresponding bytecode instruction for each operator type.
// Includes arithmetic, comparison, equality, and membership (in) operators.
func binary(p *Parser, canAssign bool) {

	opType := p.previous.Tokentype
	rule := p.getRule(opType)
	p.parsePrecedence(Precedence(rule.prec + 1))

	switch opType {
	case TOKEN_PLUS:
		p.emitByte(core.OP_ADD_NUMERIC)
	case TOKEN_PLUS_PLUS:
		p.emitByte(core.OP_ADD_VECTOR)
	case TOKEN_AMPERSAND:
		p.emitByte(core.OP_CONCAT)
	case TOKEN_MINUS:
		p.emitByte(core.OP_SUBTRACT)
	case TOKEN_STAR:
		p.emitByte(core.OP_MULTIPLY)
	case TOKEN_SLASH:
		p.emitByte(core.OP_DIVIDE)
	case TOKEN_PERCENT:
		p.emitByte(core.OP_MODULUS)
	case TOKEN_BANG_EQUAL:
		p.emitBytes(core.OP_EQUAL, core.OP_NOT)
	case TOKEN_EQUAL_EQUAL:
		p.emitByte(core.OP_EQUAL)
	case TOKEN_LESS:
		p.emitByte(core.OP_LESS)
	case TOKEN_LESS_EQUAL:
		p.emitBytes(core.OP_GREATER, core.OP_NOT)
	case TOKEN_GREATER:
		p.emitByte(core.OP_GREATER)
	case TOKEN_GREATER_EQUAL:
		p.emitBytes(core.OP_LESS, core.OP_NOT)
	case TOKEN_IN:
		p.emitByte(core.OP_IN)
	}
}

// grouping handles parenthesized expressions and tuple literals.
// For single expressions: (expr) - parses the inner expression
// For tuples: (expr1, expr2, ...) - creates a tuple with multiple values
// Automatically detects tuple syntax by looking for commas after the first expression.
func grouping(p *Parser, canAssign bool) {

	p.expression()
	if p.match(TOKEN_COMMA) {
		arity := 1
		for {
			p.expression()
			arity += 1
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
		p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after tuple.")
		p.emitByte(core.OP_CREATE_TUPLE)
		p.emitByte(uint8(arity))
	} else {
		p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
	}
}

// float parses floating-point literal tokens and emits constant bytecode.
// Converts the token's lexeme to a float64 value and adds it to the constant pool.
// Part of the prefix parsing rules for floating-point numbers.
func float(p *Parser, canAssign bool) {

	val, _ := strconv.ParseFloat(p.previous.Lexeme(), 64)
	p.emitConstant(core.MakeFloatValue(val, false))

}

// int_ parses integer literal tokens and emits constant bytecode.
// Converts the token's lexeme to an int value and adds it to the constant pool.
// Part of the prefix parsing rules for integer numbers.
func int_(p *Parser, canAssign bool) {

	val, _ := strconv.ParseInt(p.previous.Lexeme(), 10, 32)
	p.emitConstant(core.MakeIntValue(int(val), false))

}

// loxstring parses string literal tokens and emits constant bytecode.
// Removes surrounding quotes from the token lexeme and creates a string object.
// Part of the prefix parsing rules for string literals.
func loxstring(p *Parser, canAssign bool) {

	str := p.previous.Lexeme()
	str = str[1 : len(str)-1] // remove quotes

	v := core.MakeStringObjectValue(str, false)
	p.emitConstant(v)

}

// variable handles variable access expressions (identifiers).
// Delegates to namedVariable for variable resolution and bytecode generation.
// Part of the prefix parsing rules for identifier tokens.
func variable(p *Parser, canAssign bool) {

	p.namedVariable(p.previous, canAssign)
}

// unary handles unary prefix operators (- and !).
// Parses the operand expression with UNARY precedence, then emits the appropriate
// unary operation bytecode (OP_NEGATE for minus, OP_NOT for logical not).
func unary(p *Parser, canAssign bool) {

	opType := p.previous.Tokentype
	p.parsePrecedence(PREC_UNARY)

	switch opType {
	case TOKEN_MINUS:
		p.emitByte(core.OP_NEGATE)
	case TOKEN_BANG:
		p.emitByte(core.OP_NOT)
	}
}

// literal handles boolean and nil literal tokens.
// Emits the appropriate constant bytecode for true, false, and nil values.
// Part of the prefix parsing rules for literal value tokens.
func literal(p *Parser, canAssign bool) {

	switch p.previous.Tokentype {
	case TOKEN_NIL:
		p.emitByte(core.OP_NIL)
	case TOKEN_FALSE:
		p.emitByte(core.OP_FALSE)
	case TOKEN_TRUE:
		p.emitByte(core.OP_TRUE)
	}
}

// and_ handles logical AND operators with short-circuit evaluation.
// If left operand is false, jumps over right operand evaluation.
// Implements short-circuiting by conditionally evaluating the right side.
func and_(p *Parser, canAssign bool) {

	endJump := p.emitJump(core.OP_JUMP_IF_FALSE)
	p.emitByte(core.OP_POP)
	p.parsePrecedence(PREC_AND)
	p.patchJump(endJump)
}

// or_ handles logical OR operators with short-circuit evaluation.
// If left operand is true, jumps over right operand evaluation.
// Uses two jumps to implement the short-circuiting OR logic correctly.
func or_(p *Parser, canAssign bool) {

	elseJump := p.emitJump(core.OP_JUMP_IF_FALSE)
	endJump := p.emitJump(core.OP_JUMP)

	p.patchJump(elseJump)
	p.emitByte(core.OP_POP)

	p.parsePrecedence(PREC_OR)
	p.patchJump(endJump)
}

// call handles function call expressions.
// Parses the argument list and emits OP_CALL with the argument count.
// Part of the infix parsing rules for parentheses in call position.
func call(p *Parser, canAssign bool) {

	argCount := p.argumentList()
	p.emitBytes(core.OP_CALL, argCount)
}

// dot handles property access and method calls on objects.
// Supports three forms:
// 1. obj.prop = value (property assignment with OP_SET_PROPERTY)
// 2. obj.method(args) (method invocation with OP_INVOKE optimization)
// 3. obj.prop (property access with OP_GET_PROPERTY)

func dot(p *Parser, canAssign bool) {

	p.consume(TOKEN_IDENTIFIER, "Expect property name after '.'.")
	name := p.identifierConstant(p.previous)

	if p.handlePropertyCompoundAssignment(canAssign, name) {
		return
	}

	if canAssign && p.match(TOKEN_EQUAL) {
		p.expression()
		p.emitBytes(core.OP_SET_PROPERTY, name)
	} else if p.match(TOKEN_LEFT_PAREN) {
		argCount := p.argumentList()
		p.emitBytes(core.OP_INVOKE, name)
		p.emitByte(argCount)
	} else {

		p.emitBytes(core.OP_GET_PROPERTY, name)
	}
}

// handlePropertyCompoundAssignment checks for compound assignment on properties.
// e.g obj.prop += value
func (p *Parser) handlePropertyCompoundAssignment(canAssign bool, name uint8) bool {

	if canAssign && (p.check(TOKEN_PLUS_EQUAL) || p.check(TOKEN_MINUS_EQUAL)) {
		// Handle compound assignment on properties: obj.prop += value
		opType := p.current.Tokentype
		p.advance() // consume += or -=

		// Duplicate the object reference for getting the current value
		p.emitByte(core.OP_DUP)

		// Get current property value
		p.emitBytes(core.OP_GET_PROPERTY, name)

		// Parse right-hand side expression
		p.expression()

		// Perform the operation
		switch opType {
		case TOKEN_PLUS_EQUAL:
			p.emitByte(core.OP_ADD_NUMERIC)
		case TOKEN_PLUS_PLUS:
			p.emitByte(core.OP_ADD_VECTOR)
		case TOKEN_AMPERSAND:
			p.emitByte(core.OP_CONCAT)
		case TOKEN_MINUS_EQUAL:
			p.emitByte(core.OP_SUBTRACT)
		}

		// Set the property with the new value
		p.emitBytes(core.OP_SET_PROPERTY, name)
		return true
	}
	return false
}

func (p *Parser) peepHoleOptimise() {
	chunk := p.currentChunk()
	code := chunk.Code

	// Need at least 5 bytes for the pattern: GET_LOCAL(2) + GET_LOCAL(2) + ADD_NUMERIC(1)
	if len(code) < 5 {
		return
	}

	i := 0

	for i <= len(code)-5 {
		// Look for pattern: GET_LOCAL A, GET_LOCAL B, ADD_NUMERIC
		if code[i] == core.OP_GET_LOCAL &&
			code[i+2] == core.OP_GET_LOCAL &&
			code[i+4] == core.OP_ADD_NUMERIC {

			// Extract slot indices
			slotA := code[i+1]
			slotB := code[i+3]

			// Replace with optimizable superinstruction
			code[i] = core.OP_ADD_NN
			code[i+1] = slotA
			code[i+2] = slotB
			code[i+3] = 0 // specialization flag (0 = not specialized yet)
			code[i+4] = core.OP_NOOP
			i += 5
		} else {
			i++
		}
	}

}

// this handles 'this' keyword references in instance methods.
// Validates that 'this' is only used within class methods, not in functions or global scope.
// Resolves 'this' as a variable reference to access the current instance.
func this(p *Parser, canAssign bool) {
	if p.currentClass == nil {
		p.error("Can't use this outside of a class.")
		return
	}
	variable(p, false)
}

// super handles 'super' keyword for accessing superclass methods and properties.
// Validates that 'super' is only used in classes with superclasses.
// Supports both method calls (super.method(args)) and property access (super.prop).
// Uses OP_SUPER_INVOKE for method calls and OP_GET_SUPER for property access.
func super(p *Parser, canAssign bool) {

	if p.currentClass == nil {
		p.error("Cannot use 'super' outside of a class.")
	} else if !p.currentClass.hasSuperClass {
		p.error("Cannot use 'super' in a class with no superclass.")
	}

	p.consume(TOKEN_DOT, "Expect '.' after super.")
	p.consume(TOKEN_IDENTIFIER, "Expect superclass method name.")
	name := p.identifierConstant(p.previous)
	p.namedVariable(SyntheticToken("this"), false)
	if p.match(TOKEN_LEFT_PAREN) {
		argCount := p.argumentList()
		p.namedVariable(SyntheticToken("super"), false)
		p.emitBytes(core.OP_SUPER_INVOKE, name)
		p.emitByte(argCount)
	} else {
		p.namedVariable(SyntheticToken("super"), false)
		p.emitBytes(core.OP_GET_SUPER, name)
	}
}

// listLiteral handles list literal expressions [item1, item2, ...].
// Parses the list items and emits OP_CREATE_LIST with the item count.
// Part of the prefix parsing rules for square bracket tokens.
func listLiteral(p *Parser, canAssign bool) {

	listCount := p.parseList()
	p.emitBytes(core.OP_CREATE_LIST, listCount)
}

// dictLiteral handles dictionary literal expressions {key1: value1, key2: value2, ...}.
// Parses the key-value pairs and emits OP_CREATE_DICT with the pair count.
// Part of the prefix parsing rules for left brace tokens.
func dictLiteral(p *Parser, canAssign bool) {

	dictCount := p.parseDict()
	p.emitBytes(core.OP_CREATE_DICT, dictCount)
}

// slice handles indexing and slicing operations: var[expr], var[:], var[start:end], etc.
// Supports Python-style slicing with various forms:
// - [expr] for single element indexing
// - [:] for full slice
// - [:end] for slice from beginning
// - [start:] for slice to end
// - [start:end] for range slice
// Delegates to specific slice functions based on the syntax pattern detected.
func slice(p *Parser, canAssign bool) {

	if p.check(TOKEN_RIGHT_BRACKET) {
		p.error("Can't have empty slice.")
		return
	}

	_ = p.identifierConstant(p.previous)

	// handle the slice variants : [exp], [:], [:exp], [exp:], [exp:exp]
	if p.match(TOKEN_COLON) {
		//[:],[:exp]
		p.slice1(canAssign)

	} else {
		// [exp],[exp:],[exp:exp]
		// slice from/index -> stack
		p.expression()
		if p.match(TOKEN_RIGHT_BRACKET) {
			//[exp]
			p.index(canAssign)

		} else {
			// [exp:],[exp:exp]
			if p.match(TOKEN_COLON) {
				p.slice2(canAssign)

			}
		}
	}
}

// str_ handles str() function calls for type conversion to strings.
// Parses the expression inside parentheses and emits OP_STR for string conversion.
// Part of the prefix parsing rules for 'str' keyword followed by parentheses.
func str_(p *Parser, canAssign bool) {
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after str.")
	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
	p.emitByte(core.OP_STR)
}
