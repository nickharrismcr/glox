package main

import (
	"fmt"
	"strconv"
	"strings"
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
	PREC_FACTOR                // * /
	PREC_UNARY                 // ! -
	PREC_CALL                  // . ()
	PREC_PRIMARY
)

type ParseFn func(*Parser, bool)

type ParseRule struct {
	prefix ParseFn
	infix  ParseFn
	prec   Precedence
}

type Local struct {
	name  Token
	depth int
}

type Loop struct {
	parent *Loop
	start  int
	break_ int
}

func NewLoop() *Loop {
	return &Loop{}
}

type FunctionType int

const (
	TYPE_FUNCTION FunctionType = iota
	TYPE_SCRIPT
)

type Compiler struct {
	enclosing  *Compiler
	function   *FunctionObject
	type_      FunctionType
	locals     [256]Local
	localCount int
	scopeDepth int
	loop       *Loop
}

func NewCompiler(type_ FunctionType, parent *Compiler) *Compiler {

	rv := &Compiler{
		enclosing: parent,
		type_:     type_,
		function:  makeFunctionObject(),
	}
	// slot 0 is for enclosing function
	rv.locals[0] = Local{
		depth: 0,
		name:  Token{},
	}
	rv.localCount = 1
	return rv
}

type Parser struct {
	scanner             *Scanner
	current, previous   Token
	hadError, panicMode bool
	rules               map[TokenType]ParseRule
	currentCompiler     *Compiler
}

func NewParser() *Parser {
	p := &Parser{
		hadError:  false,
		panicMode: false,
	}
	p.setRules()
	return p
}

func (vm *VM) compile(source string) *FunctionObject {

	if debugTraceExecution {
		fmt.Println("Compiling...")
	}
	parser := NewParser()
	parser.scanner = NewScanner(source)
	parser.currentCompiler = NewCompiler(TYPE_SCRIPT, nil)
	parser.advance()
	for !parser.match(TOKEN_EOF) {
		parser.declaration()
	}
	parser.consume(TOKEN_EOF, "Expect end of expression")
	function := parser.endCompiler()
	if debugTraceExecution {
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
		TOKEN_LEFT_BRACE:    {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_RIGHT_BRACE:   {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_COMMA:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_DOT:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_MINUS:         {prefix: unary, infix: binary, prec: PREC_TERM},
		TOKEN_PLUS:          {prefix: nil, infix: binary, prec: PREC_TERM},
		TOKEN_SEMICOLON:     {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_SLASH:         {prefix: nil, infix: binary, prec: PREC_FACTOR},
		TOKEN_STAR:          {prefix: nil, infix: binary, prec: PREC_FACTOR},
		TOKEN_BANG:          {prefix: unary, infix: nil, prec: PREC_NONE},
		TOKEN_BANG_EQUAL:    {prefix: nil, infix: binary, prec: PREC_EQUALITY},
		TOKEN_EQUAL:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {prefix: nil, infix: binary, prec: PREC_EQUALITY},
		TOKEN_GREATER:       {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_GREATER_EQUAL: {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_LESS:          {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_LESS_EQUAL:    {prefix: nil, infix: binary, prec: PREC_COMPARISON},
		TOKEN_IDENTIFIER:    {prefix: variable, infix: nil, prec: PREC_NONE},
		TOKEN_STRING:        {prefix: loxstring, infix: nil, prec: PREC_NONE},
		TOKEN_NUMBER:        {prefix: number, infix: nil, prec: PREC_NONE},
		TOKEN_AND:           {prefix: nil, infix: and_, prec: PREC_AND},
		TOKEN_CLASS:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_ELSE:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FALSE:         {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_FOR:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FUNC:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_IF:            {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_NIL:           {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_OR:            {prefix: nil, infix: or_, prec: PREC_OR},
		TOKEN_PRINT:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_RETURN:        {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_SUPER:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_THIS:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_TRUE:          {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_VAR:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_CONST:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_WHILE:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_ERROR:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_EOF:           {prefix: nil, infix: nil, prec: PREC_NONE},
	}
}

func (p *Parser) match(tt TokenType) bool {

	if !p.check(tt) {
		return false
	}
	p.advance()
	return true
}

func (p *Parser) check(tt TokenType) bool {
	return p.current.tokentype == tt
}

func (p *Parser) advance() {

	p.previous = p.current
	for {
		p.current = p.scanner.scanToken()
		if p.current.tokentype != TOKEN_ERROR {
			break
		}
		p.errorAtCurrent(p.current.lexeme())
	}

}

func (p *Parser) getRule(tok TokenType) ParseRule {
	return p.rules[tok]
}

func (p *Parser) declaration() {
	if p.match(TOKEN_FUNC) {
		p.funcDeclaration()
	} else if p.match(TOKEN_VAR) {
		p.varDeclaration()
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
	} else if p.match(TOKEN_BREAK) {
		p.breakStatement()
	} else if p.match(TOKEN_CONTINUE) {
		p.continueStatement()
	} else if p.match(TOKEN_FOR) {
		p.forStatement()
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

func (p *Parser) expression() {
	p.parsePredence(PREC_ASSIGNMENT)
}

func (p *Parser) block() {

	for !p.check(TOKEN_RIGHT_BRACE) && !p.check(TOKEN_EOF) {
		p.declaration()
	}
	p.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (p *Parser) funcDeclaration() {
	global := p.parseVariable("Expect function name.")
	p.markInitialised()
	p.function(TYPE_FUNCTION)
	p.defineVariable(global)
}

func (p *Parser) function(type_ FunctionType) {

	p.currentCompiler = NewCompiler(type_, p.currentCompiler)
	funcname := p.previous.lexeme()
	p.currentCompiler.function.name = MakeStringObject(funcname)
	p.beginScope()

	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after function name.")
	if !p.check(TOKEN_RIGHT_PAREN) {
		for {
			p.currentCompiler.function.arity++
			if p.currentCompiler.function.arity > 255 {
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
	p.consume(TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	p.block()

	function := p.endCompiler()
	p.emitBytes(OP_CONSTANT, p.makeConstant(makeObjectValue(function, false)))
}

func (p *Parser) varDeclaration() {

	global := p.parseVariable("Expect variable name")

	if p.match(TOKEN_EQUAL) {
		p.expression()
	} else {
		p.emitByte(OP_NIL)
	}
	p.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration")

	p.defineVariable(global)
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
	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
	p.emitByte(OP_POP)
}

func (p *Parser) ifStatement() {
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect '(' after condition.")

	thenJump := p.emitJump(OP_JUMP_IF_FALSE)
	p.emitByte(OP_POP)
	p.statement()
	elseJump := p.emitJump(OP_JUMP)
	p.patchJump(thenJump)
	p.emitByte(OP_POP)
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
		p.expression()
		p.consume(TOKEN_SEMICOLON, "Expect ';' after return value.")
		p.emitByte(OP_RETURN)
	}
}

func (p *Parser) whileStatement() {

	loopSave := p.currentCompiler.loop
	p.currentCompiler.loop = NewLoop()

	p.currentCompiler.loop.start = len(p.currentChunk().code)
	p.consume(TOKEN_LEFT_PAREN, "Expect '(' after while.")
	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := p.emitJump(OP_JUMP_IF_FALSE)
	p.emitByte(OP_POP)
	p.statement()
	if p.currentCompiler.loop.break_ != 0 {
		p.patchJump(p.currentCompiler.loop.break_)
	}
	p.emitLoop(p.currentCompiler.loop.start)
	p.patchJump(exitJump)
	p.emitByte(OP_POP)

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
		p.varDeclaration()
	} else {
		p.expressionStatement()
	}
	p.currentCompiler.loop.start = len(p.currentChunk().code)
	// exit condition
	exitJump := -1
	if !p.match(TOKEN_SEMICOLON) {
		p.expression()
		p.consume(TOKEN_SEMICOLON, "Expect ';'.")
		exitJump = p.emitJump(OP_JUMP_IF_FALSE)
		p.emitByte(OP_POP)
	}
	// increment
	if !p.match(TOKEN_RIGHT_PAREN) {
		// jump over increment, will be executed after body
		bodyJump := p.emitJump(OP_JUMP)
		incrementStart := len(p.currentChunk().code)
		p.expression()
		p.emitByte(OP_POP)
		p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for clauses.")
		p.emitLoop(p.currentCompiler.loop.start)
		p.currentCompiler.loop.start = incrementStart
		p.patchJump(bodyJump)
	}
	p.statement()
	if p.currentCompiler.loop.break_ != 0 {
		p.patchJump(p.currentCompiler.loop.break_)
	}
	p.emitLoop(p.currentCompiler.loop.start)

	if exitJump != -1 {
		p.patchJump(exitJump)
		p.emitByte(OP_POP)
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
			p.emitByte(OP_POP)
		}
	}
	p.currentCompiler.loop.break_ = p.emitJump(OP_JUMP)
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
			p.emitByte(OP_POP)
		}
	}
	p.emitLoop(p.currentCompiler.loop.start)
}

func (p *Parser) printStatement() {
	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	p.emitByte(OP_PRINT)
}
func (p *Parser) synchronize() {
	p.panicMode = false
	for p.current.tokentype != TOKEN_EOF {
		if p.previous.tokentype == TOKEN_SEMICOLON {
			return
		}
		switch p.current.tokentype {
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

	if p.current.tokentype == toktype {
		p.advance()
		return
	}
	p.errorAtCurrent(msg)
}

func (p *Parser) emitByte(byte uint8) {
	p.currentChunk().writeOpCode(byte, p.previous.line)
}

func (p *Parser) emitBytes(byte1, byte2 uint8) {
	p.emitByte(byte1)
	p.emitByte(byte2)
}

func (p *Parser) emitLoop(loopStart int) {
	p.emitByte(OP_LOOP)

	offset := len(p.currentChunk().code) - loopStart + 2
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
	return len(p.currentChunk().code) - 2
}

func (p *Parser) currentChunk() *Chunk {
	return p.currentCompiler.function.chunk
}

func (p *Parser) endCompiler() *FunctionObject {
	p.emitReturn()
	function := p.currentCompiler.function
	if debugPrintCode {
		if !p.hadError {
			s := ""
			if function.name.String() == "" {
				s = "<script>"
			} else {
				s = function.name.String()
			}
			p.currentChunk().disassemble(s)
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
		p.emitByte(OP_POP)
		c.localCount--
	}
}

func (p *Parser) parsePredence(prec Precedence) {

	p.advance()

	prefixRule := p.getRule(p.previous.tokentype).prefix
	if prefixRule == nil {
		p.error("Expect expression")
		return
	}

	canAssign := prec <= PREC_ASSIGNMENT
	prefixRule(p, canAssign)
	for prec <= p.getRule(p.current.tokentype).prec {

		p.advance()
		infixRule := p.getRule(p.previous.tokentype).infix
		if infixRule != nil {

			infixRule(p, canAssign)
		}

	}
	if canAssign && p.match(TOKEN_EQUAL) {
		p.error("Invalid assignment target.")
	}
}

func (p *Parser) identifierConstant(t Token) uint8 {
	return p.makeConstant(makeObjectValue(MakeStringObject(t.lexeme()), false))
}

func (p *Parser) identifiersEqual(a, b Token) bool {
	if a.length != b.length {
		return false
	}
	if a.lexeme() != b.lexeme() {
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
	c.constants[len(c.constants)-1] = immutable(c.constants[len(c.constants)-1])
}

func (p *Parser) defineVariable(global uint8) {
	// if local, it will already be on the stack
	if p.currentCompiler.scopeDepth > 0 {
		p.markInitialised()
		return
	}
	p.emitBytes(OP_DEFINE_GLOBAL, global)
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

func (p *Parser) defineConstVariable(global uint8) {
	// if local, it will already be on the stack
	if p.currentCompiler.scopeDepth > 0 {
		p.markInitialised()
		p.setLocalImmutable()
		return
	}
	p.emitBytes(OP_DEFINE_GLOBAL_CONST, global)
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

func (p *Parser) namedVariable(name Token, canAssign bool) {

	var getOp, setOp uint8

	arg := p.resolveLocal(p.currentCompiler, name)
	if arg != -1 {
		getOp = OP_GET_LOCAL
		setOp = OP_SET_LOCAL
	} else {
		arg = int(p.identifierConstant(name))
		getOp = OP_GET_GLOBAL
		setOp = OP_SET_GLOBAL
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
	local := Local{
		name:  name,
		depth: -1, // marks as uninitialised
	}
	p.currentCompiler.locals[p.currentCompiler.localCount] = local
	p.currentCompiler.localCount++
}

func (p *Parser) emitConstant(value Value) {
	p.emitBytes(OP_CONSTANT, p.makeConstant(value))
}

func (p *Parser) patchJump(offset int) {

	jump := len(p.currentChunk().code) - offset - 2
	if uint16(jump) > ^uint16(0) {
		p.error("Jump overflow")
	}
	p.currentChunk().code[offset] = uint8((jump >> 8) & 0xff)
	p.currentChunk().code[offset+1] = uint8(jump & 0xff)

}

func (p *Parser) makeConstant(value Value) uint8 {
	constidx := p.currentChunk().addConstant(value)
	if constidx > 255 {
		p.error("Too many constants in one chunk")
		return 0
	}
	return constidx
}

func (p *Parser) emitReturn() {
	p.emitByte(OP_NIL)
	p.emitByte(OP_RETURN)
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
	fmt.Printf("[line %d] Error ", tok.line)
	if tok.tokentype == TOKEN_EOF {
		fmt.Printf(" at end")
	} else if tok.tokentype == TOKEN_ERROR {
		fmt.Printf(" at %s ", tok.lexeme())
	} else {
		fmt.Printf(" at %s ", tok.lexeme())
	}
	fmt.Printf(" : %s\n", msg)
	p.hadError = true
}

//=============================================================================
// pratt parser functions

func binary(p *Parser, canAssign bool) {

	opType := p.previous.tokentype
	rule := p.getRule(opType)
	p.parsePredence(Precedence(rule.prec + 1))

	switch opType {
	case TOKEN_PLUS:
		p.emitByte(OP_ADD)
	case TOKEN_MINUS:
		p.emitByte(OP_SUBTRACT)
	case TOKEN_STAR:
		p.emitByte(OP_MULTIPLY)
	case TOKEN_SLASH:
		p.emitByte(OP_DIVIDE)
	case TOKEN_BANG_EQUAL:
		p.emitBytes(OP_EQUAL, OP_NOT)
	case TOKEN_EQUAL_EQUAL:
		p.emitByte(OP_EQUAL)
	case TOKEN_LESS:
		p.emitByte(OP_LESS)
	case TOKEN_LESS_EQUAL:
		p.emitBytes(OP_GREATER, OP_NOT)
	case TOKEN_GREATER:
		p.emitByte(OP_GREATER)
	case TOKEN_GREATER_EQUAL:
		p.emitBytes(OP_LESS, OP_NOT)
	}
}

func grouping(p *Parser, canAssign bool) {

	p.expression()
	p.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func number(p *Parser, canAssign bool) {

	val, _ := strconv.ParseFloat(p.previous.lexeme(), 64)
	p.emitConstant(makeNumberValue(val, false))

}

func loxstring(p *Parser, canAssign bool) {

	str := p.previous.lexeme()
	strobj := MakeStringObject(strings.Replace(str, "\"", "", -1))
	p.emitConstant(makeObjectValue(&strobj, false))

}

func variable(p *Parser, canAssign bool) {
	p.namedVariable(p.previous, canAssign)
}

func unary(p *Parser, canAssign bool) {

	opType := p.previous.tokentype
	p.parsePredence(PREC_UNARY)

	switch opType {
	case TOKEN_MINUS:
		p.emitByte(OP_NEGATE)
	case TOKEN_BANG:
		p.emitByte(OP_NOT)
	}
}

func literal(p *Parser, canAssign bool) {
	switch p.previous.tokentype {
	case TOKEN_NIL:
		p.emitByte(OP_NIL)
	case TOKEN_FALSE:
		p.emitByte(OP_FALSE)
	case TOKEN_TRUE:
		p.emitByte(OP_TRUE)
	}
}

func and_(p *Parser, canAssign bool) {

	endJump := p.emitJump(OP_JUMP_IF_FALSE)
	p.emitByte(OP_POP)
	p.parsePredence(PREC_AND)
	p.patchJump(endJump)
}

func or_(p *Parser, canAssign bool) {

	elseJump := p.emitJump(OP_JUMP_IF_FALSE)
	endJump := p.emitJump(OP_JUMP)

	p.patchJump(elseJump)
	p.emitByte(OP_POP)

	p.parsePredence(PREC_OR)
	p.patchJump(endJump)
}

func call(p *Parser, canAssign bool) {

	argCount := p.argumentList()
	p.emitBytes(OP_CALL, argCount)
}
