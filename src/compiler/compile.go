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

func NewParser() *Parser {

	p := &Parser{
		hadError:  false,
		panicMode: false,
		globals:   map[string]bool{},
	}
	p.setRules()
	return p
}

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

func (p *Parser) match(tt TokenType) bool {

	if !p.check(tt) {
		return false
	}
	if tt != TOKEN_EOF {
		p.advance()
	}
	return true
}

func (p *Parser) check(tt TokenType) bool {

	return p.current.Tokentype == tt
}

func (p *Parser) checkNext(tt TokenType) bool {

	return p.scn.Tokens.At(p.scn.TokenIdx+1).Tokentype == tt
}

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

func (p *Parser) getRule(tok TokenType) ParseRule {

	return p.rules[tok]
}

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
func (p *Parser) statement() {

	if p.match(TOKEN_PRINT) {
		p.printStatement()
	} else if p.match(TOKEN_IMPORT) {
		p.importStatement()
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

func (p *Parser) raiseStatement() {

	p.expression() // this includes constructor calls
	p.consume(TOKEN_SEMICOLON, "Expect ';' after throw expression.")
	p.emitByte(core.OP_RAISE)
}

func (p *Parser) importStatement() {
	c := 0
	for {
		p.consume(TOKEN_IDENTIFIER, "Expect module name.")
		nameConstant := p.identifierConstant(p.previous)
		c = c + 1
		p.emitBytes(core.OP_IMPORT, nameConstant)
		if !p.match(TOKEN_COMMA) {
			break
		}
	}
	p.consume(TOKEN_SEMICOLON, "Expect ';' after import list.")
}
func (p *Parser) expression() {

	p.parsePredence(PREC_ASSIGNMENT)
}

func (p *Parser) block() {

	for !p.check(TOKEN_RIGHT_BRACE) && !p.check(TOKEN_EOF) {
		p.declaration()
	}
	p.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
	p.match(TOKEN_EOL) // allow EOL after block

}

func (p *Parser) funcDeclaration() {

	global := p.parseVariable("Expect function name.")
	p.markInitialised()
	p.function(TYPE_FUNCTION)
	p.defineVariable(global)
}

func (p *Parser) function(type_ FunctionType) {

	compiler := NewCompiler(type_, p.currentCompiler.scriptName, p.currentCompiler, p.currentCompiler.environment)
	p.currentCompiler = compiler
	funcname := p.previous.Lexeme()

	compiler.function.Name = core.MakeStringObject(funcname)

	p.beginScope()

	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after function name.")
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			p.currentCompiler.function.Arity++
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
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after function parameters.")
	p.match(TOKEN_EOL) // allow EOL after parameters
	p.consume(TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	p.block()

	function := p.endCompiler()
	p.emitBytes(core.OP_CLOSURE, p.MakeConstant(core.MakeObjectValue(function, false)))

	for i := 0; i < function.UpvalueCount; i++ {
		uv := *(compiler.upvalues[i])
		if uv.isLocal {
			p.emitByte(1)
		} else {
			p.emitByte(0)
		}
		p.emitByte(uv.index)
	}
}

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

func (p *Parser) expressionStatement() {

	//handle implicit declarations
	if p.check(TOKEN_IDENTIFIER) && p.checkNext(TOKEN_EQUAL) {
		name := p.current
		p.consume(TOKEN_IDENTIFIER, "")
		p.consume(TOKEN_EQUAL, "")
		l := name.Lexeme()
		if p.currentCompiler.scopeDepth > 0 {
			if p.resolveLocal(p.currentCompiler, name) == -1 &&
				p.resolveUpvalue(p.currentCompiler, name) == -1 &&
				!p.checkGlobals(l) {
				p.addLocal(name)
				p.currentCompiler.locals[p.currentCompiler.localCount-1].depth = p.currentCompiler.scopeDepth
				p.expression()
				p.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
			}
		} else {
			if _, ok := p.globals[l]; !ok {
				p.globals[l] = true
				p.expression()
				p.emitBytes(core.OP_DEFINE_GLOBAL, p.identifierConstant(name))
			}
		}
		return
	}

	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
	p.emitByte(core.OP_POP)
}

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

func (p *Parser) breakStatement() {

	p.consume(TOKEN_SEMICOLON, "Expect ';' after statement.")
	if p.currentCompiler.loop == nil {
		p.errorAtCurrent("Cannot use break outside loop.")
	}

	// drop local vars on stack
	c := p.currentCompiler

	for i := 0; i < c.localCount; i++ {
		if c.locals[i].depth >= c.scopeDepth-1 {
			p.emitByte(core.OP_POP)
		}
	}
	p.currentCompiler.loop.breaks = append(p.currentCompiler.loop.breaks, p.emitJump(core.OP_JUMP))
}

func (p *Parser) breakpointStatement() {

	p.consume(TOKEN_SEMICOLON, "Expect ';' after statement.")
	p.emitByte(core.OP_BREAKPOINT)
}

func (p *Parser) continueStatement() {

	p.consume(TOKEN_SEMICOLON, "Expect ';' after statement.")
	if p.currentCompiler.loop == nil {
		p.errorAtCurrent("Cannot use continue outside loop.")
	}

	// drop local vars on stack
	c := p.currentCompiler
	for i := 0; i < c.localCount; i++ {
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

// creates 3 locals on stack:
//   - var receiving iterator output
//   - iterated list/string
//   - iteration index
func (p *Parser) foreachStatement() {

	loopSave := p.currentCompiler.loop
	p.currentCompiler.loop = NewLoop()
	p.currentCompiler.loop.foreach = true // so continue knows to jump to next

	p.beginScope()
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after for.")
	p.consume(TOKEN_VAR, "Expect var declaration")
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

func (p *Parser) printStatement() {

	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	p.emitByte(core.OP_STR)
	p.emitByte(core.OP_PRINT)
}
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

func (p *Parser) consume(toktype TokenType, msg string) {

	if p.current.Tokentype == toktype || (toktype == TOKEN_SEMICOLON && p.current.Tokentype == TOKEN_EOL) {
		p.advance()
		return
	}
	p.errorAtCurrent(msg)
}

func (p *Parser) emitByte(byte uint8) {

	p.currentChunk().WriteOpCode(byte, p.previous.Line)
}

func (p *Parser) emitBytes(byte1, byte2 uint8) {

	p.emitByte(byte1)
	p.emitByte(byte2)
}

func (p *Parser) emitLoop(instr uint8, loopStart int) {

	p.emitByte(instr)

	offset := len(p.currentChunk().Code) - loopStart + 2
	if offset >= int(^uint16(0)) {
		p.error("Loop body too large")
	}

	p.emitByte(uint8((offset >> 8) & 0xff))
	p.emitByte(uint8(offset & 0xff))
}

func (p *Parser) emitJump(instr uint8) int {

	p.emitByte(instr)
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.currentChunk().Code) - 2
}
func (p *Parser) emitForeach(slot uint8, iterslot uint8) int {

	p.emitByte(core.OP_FOREACH)
	p.emitByte(slot)
	p.emitByte(iterslot)
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.currentChunk().Code) - 3
}

func (p *Parser) emitTry() int {

	p.emitByte(core.OP_TRY)
	p.emitByte(0xff)
	p.emitByte(0xff)
	return len(p.currentChunk().Code) - 2
}

func (p *Parser) currentChunk() *core.Chunk {

	return p.currentCompiler.function.Chunk
}

func (p *Parser) endCompiler() *core.FunctionObject {

	p.emitReturn()

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

func (p *Parser) beginScope() {

	p.currentCompiler.scopeDepth++
}

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
		c.localCount--
	}
}

func (p *Parser) parsePredence(prec Precedence) {

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

func (p *Parser) identifierConstant(t Token) uint8 {

	x := t.Lexeme()
	p.globals[x] = true
	return p.MakeConstant(core.MakeObjectValue(core.MakeStringObject(t.Lexeme()), false))
}

func (p *Parser) identifiersEqual(a, b Token) bool {

	if a.Length != b.Length {
		return false
	}
	if a.Lexeme() != b.Lexeme() {
		return false
	}
	return true
}

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
func (p *Parser) addUpvalue(compiler *Compiler, index uint8, isLocal bool) int {

	upvalueCount := compiler.function.UpvalueCount

	// does upvalue already exist ?
	for i := 0; i < upvalueCount; i++ {
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
	compiler.function.UpvalueCount++

	return upvalueCount

}
func (p *Parser) resolveUpvalue(compiler *Compiler, name Token) int {

	/*
		First, we look for a matching local variable in the enclosing function.
		If we find one, we capture that local and return. That’s the base case.
		Otherwise, we look for a local variable beyond the immediately enclosing function.
		We do that by recursively calling resolveUpvalue() on the enclosing compiler, not the current one.
		This series of resolveUpvalue() calls works its way along the chain of nested compilers until it hits
		one of the base cases—either it finds an actual local variable to capture or it runs out of compilers.

		When a local variable is found, the most deeply nested call to resolveUpvalue() captures it and returns the upvalue index.
		That returns to the next call for the inner function declaration. That call captures the upvalue from the
		surrounding function, and so on. As each nested call to resolveUpvalue() returns, we drill back down into
		the innermost function declaration where the identifier we are resolving appears. At each step along
		the way, we add an upvalue to the intervening function and pass the resulting upvalue index down to the next call.

		Note that the new call to addUpvalue() passes false for the isLocal parameter. Now you see that that flag controls whether
		the closure captures a local variable or an upvalue from the surrounding function.
	*/

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

func (p *Parser) parseVariable(errorMsg string) uint8 {

	p.consume(TOKEN_IDENTIFIER, errorMsg)
	p.declareVariable()
	// if local, don't add to constant table
	if p.currentCompiler.scopeDepth > 0 {
		return 0
	}
	return p.identifierConstant(p.previous)
}

func (p *Parser) markInitialised() {

	c := p.currentCompiler
	if c.scopeDepth == 0 {
		return
	}
	c.locals[c.localCount-1].depth = c.scopeDepth
}

func (p *Parser) setLocalImmutable() {

	c := p.currentChunk()
	c.Constants[len(c.Constants)-1] = core.Immutable(c.Constants[len(c.Constants)-1])
}

func (p *Parser) defineVariable(global uint8) {

	// if local, it will already be on the stack
	if p.currentCompiler.scopeDepth > 0 {
		p.markInitialised()
		return
	}
	p.emitBytes(core.OP_DEFINE_GLOBAL, global)
}

func (p *Parser) argumentList() uint8 {

	var argCount uint8 = 0
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			p.expression()
			argCount++
			if argCount == 255 {
				p.error("Can't have more than 255 arguments. ")
			}
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after arguments")
	return argCount
}

func (p *Parser) parseList() uint8 {

	var itemCount uint8 = 0
	if !p.check(TOKEN_RIGHT_BRACKET) {
		for {
			p.expression()
			itemCount++
			if itemCount == 255 {
				p.error("Can't have more than 255 initialiser items. ")
			}
			if !p.match(TOKEN_COMMA) {
				break
			}
		}
	}
	p.consume(TOKEN_RIGHT_BRACKET, "Expect ']' after list items.")
	return itemCount
}

func (p *Parser) parseDict() uint8 {

	var itemCount uint8 = 0
	if !p.match(TOKEN_RIGHT_BRACE) {
		for {
			p.expression()
			p.consume(TOKEN_COLON, "Expect ':' after key.")
			p.expression()
			itemCount++
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

func (p *Parser) defineConstVariable(global uint8) {

	// if local, it will already be on the stack
	if p.currentCompiler.scopeDepth > 0 {
		p.markInitialised()
		p.setLocalImmutable()
		return
	}
	p.emitBytes(core.OP_DEFINE_GLOBAL_CONST, global)
}

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

func (p *Parser) checkGlobals(name string) bool {
	_, ok := p.globals[name]
	return ok
}

func (p *Parser) namedVariable(name Token, canAssign bool) {

	var getOp, setOp uint8
	a := name.Lexeme()
	_ = a
	arg := p.resolveLocal(p.currentCompiler, name)
	if arg != -1 {
		getOp = core.OP_GET_LOCAL
		setOp = core.OP_SET_LOCAL
	} else if arg = p.resolveUpvalue(p.currentCompiler, name); arg != -1 {
		getOp = core.OP_GET_UPVALUE
		setOp = core.OP_SET_UPVALUE
	} else {
		arg = int(p.identifierConstant(name))
		getOp = core.OP_GET_GLOBAL
		setOp = core.OP_SET_GLOBAL
	}

	if canAssign && p.match(TOKEN_EQUAL) {
		p.expression()
		p.emitBytes(setOp, uint8(arg))
	} else {
		p.emitBytes(getOp, uint8(arg))
	}

}

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
	p.currentCompiler.localCount++
}

func (p *Parser) emitConstant(value core.Value) {

	p.emitBytes(core.OP_CONSTANT, p.MakeConstant(value))
}

func (p *Parser) patchJump(offset int) {

	jump := len(p.currentChunk().Code) - offset - 2
	if uint16(jump) > ^uint16(0) {
		p.error("Jump overflow")
	}
	p.currentChunk().Code[offset] = uint8((jump >> 8) & 0xff)
	p.currentChunk().Code[offset+1] = uint8(jump & 0xff)

}
func (p *Parser) patchForeach(offset int) {

	jump := len(p.currentChunk().Code) - offset - 2
	if uint16(jump) > ^uint16(0) {
		p.error("Jump overflow")
	}
	p.currentChunk().Code[offset+1] = uint8((jump >> 8) & 0xff)
	p.currentChunk().Code[offset+2] = uint8(jump & 0xff)

}

func (p *Parser) patchTry(offset int) {

	address := len(p.currentChunk().Code)
	p.currentChunk().Code[offset] = uint8((address >> 8) & 0xff)
	p.currentChunk().Code[offset+1] = uint8(address & 0xff)
}

func (p *Parser) MakeConstant(value core.Value) uint8 {

	constidx := p.currentChunk().AddConstant(value)
	if constidx > 254 {
		p.error("Too many constants in one chunk")
		return 0
	}
	return constidx
}

func (p *Parser) emitReturn() {

	if p.currentCompiler.type_ == TYPE_INITIALIZER {
		p.emitBytes(core.OP_GET_LOCAL, 0)
	} else {
		p.emitByte(core.OP_NIL)
	}
	op := core.OP_RETURN

	p.emitByte(op)
}

// a[:], a[:exp]
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

// a[exp]
func (p *Parser) index(canAssign bool) {

	if canAssign && p.match(TOKEN_EQUAL) {
		// RHS -> stack
		p.expression()
		p.emitByte(core.OP_INDEX_ASSIGN)
	} else {
		p.emitByte(core.OP_INDEX)
	}
}

// a[exp:], a[exp:exp]
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

func (p *Parser) errorAtCurrent(msg string) {

	p.errorAt(p.current, msg)
}

func (p *Parser) error(msg string) {

	p.errorAt(p.previous, msg)
}

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

func binary(p *Parser, canAssign bool) {

	opType := p.previous.Tokentype
	rule := p.getRule(opType)
	p.parsePredence(Precedence(rule.prec + 1))

	switch opType {
	case TOKEN_PLUS:
		p.emitByte(core.OP_ADD)
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

func grouping(p *Parser, canAssign bool) {

	p.expression()
	if p.match(TOKEN_COMMA) {
		arity := 1
		for {
			p.expression()
			arity++
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

func float(p *Parser, canAssign bool) {

	val, _ := strconv.ParseFloat(p.previous.Lexeme(), 64)
	p.emitConstant(core.MakeFloatValue(val, false))

}

func int_(p *Parser, canAssign bool) {

	val, _ := strconv.ParseInt(p.previous.Lexeme(), 10, 32)
	p.emitConstant(core.MakeIntValue(int(val), false))

}

func loxstring(p *Parser, canAssign bool) {

	str := p.previous.Lexeme()
	str = str[1 : len(str)-1] // remove quotes

	strobj := core.MakeStringObject(str)
	p.emitConstant(core.MakeObjectValue(strobj, false))

}

func variable(p *Parser, canAssign bool) {

	p.namedVariable(p.previous, canAssign)
}

func unary(p *Parser, canAssign bool) {

	opType := p.previous.Tokentype
	p.parsePredence(PREC_UNARY)

	switch opType {
	case TOKEN_MINUS:
		p.emitByte(core.OP_NEGATE)
	case TOKEN_BANG:
		p.emitByte(core.OP_NOT)
	}
}

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

func and_(p *Parser, canAssign bool) {

	endJump := p.emitJump(core.OP_JUMP_IF_FALSE)
	p.emitByte(core.OP_POP)
	p.parsePredence(PREC_AND)
	p.patchJump(endJump)
}

func or_(p *Parser, canAssign bool) {

	elseJump := p.emitJump(core.OP_JUMP_IF_FALSE)
	endJump := p.emitJump(core.OP_JUMP)

	p.patchJump(elseJump)
	p.emitByte(core.OP_POP)

	p.parsePredence(PREC_OR)
	p.patchJump(endJump)
}

func call(p *Parser, canAssign bool) {

	argCount := p.argumentList()
	p.emitBytes(core.OP_CALL, argCount)
}

func dot(p *Parser, canAssign bool) {

	p.consume(TOKEN_IDENTIFIER, "Expect property name after '.'.")
	name := p.identifierConstant(p.previous)

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

func this(p *Parser, canAssign bool) {
	if p.currentClass == nil {
		p.error("Can't use this outside of a class.")
		return
	}
	variable(p, false)
}

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

func listLiteral(p *Parser, canAssign bool) {

	listCount := p.parseList()
	p.emitBytes(core.OP_CREATE_LIST, listCount)
}

func dictLiteral(p *Parser, canAssign bool) {

	dictCount := p.parseDict()
	p.emitBytes(core.OP_CREATE_DICT, dictCount)
}

// var[<expr>]
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

func str_(p *Parser, canAssign bool) {
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after str.")
	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
	p.emitByte(core.OP_STR)
}
