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

type Parser struct {
	scanner             *Scanner
	compilingChunk      *Chunk
	current, previous   Token
	hadError, panicMode bool
	rules               map[TokenType]ParseRule
}

func NewParser() *Parser {
	p := &Parser{
		hadError:  false,
		panicMode: false,
	}
	p.setRules()
	return p
}

func (vm *VM) compile(source string) bool {

	if debugTraceExecution {
		fmt.Println("Compiling...")
	}
	parser := NewParser()
	parser.compilingChunk = vm.chunk
	parser.scanner = NewScanner(source)
	parser.advance()
	for !parser.match(TOKEN_EOF) {
		parser.declaration()
	}
	parser.consume(TOKEN_EOF, "Expect end of expression")
	parser.endCompiler()
	if debugTraceExecution {
		fmt.Println("Compile done.")
	}
	return !parser.hadError
}

func (p *Parser) setRules() {

	p.rules = map[TokenType]ParseRule{
		TOKEN_LEFT_PAREN:    {prefix: grouping, infix: nil, prec: PREC_NONE},
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
		TOKEN_AND:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_CLASS:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_ELSE:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FALSE:         {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_FOR:           {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_FUNC:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_IF:            {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_NIL:           {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_OR:            {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_PRINT:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_RETURN:        {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_SUPER:         {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_THIS:          {prefix: nil, infix: nil, prec: PREC_NONE},
		TOKEN_TRUE:          {prefix: literal, infix: nil, prec: PREC_NONE},
		TOKEN_VAR:           {prefix: nil, infix: nil, prec: PREC_NONE},
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
		if debugTraceExecution {
			fmt.Printf("Lexeme : %s\n", p.current.lexeme())
		}
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
	if p.match(TOKEN_VAR) {
		p.varDeclaration()
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
	} else {
		p.expressionStatement()
	}
}

func (p *Parser) expression() {
	p.parsePredence(PREC_ASSIGNMENT)
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

func (p *Parser) expressionStatement() {
	p.expression()
	p.consume(TOKEN_SEMICOLON, "Expect ';' after expression.")
	p.emitByte(OP_POP)
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

func (p *Parser) currentChunk() *Chunk {
	return p.compilingChunk
}

func (p *Parser) endCompiler() {
	p.emitReturn()
	if debugPrintCode {
		if !p.hadError {
			p.compilingChunk.disassemble("code")
		}
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
	return p.makeConstant(MakeObjectValue(MakeStringObject(t.lexeme())))
}

func (p *Parser) parseVariable(errorMsg string) uint8 {
	p.consume(TOKEN_IDENTIFIER, errorMsg)
	return p.identifierConstant(p.previous)
}

func (p *Parser) defineVariable(global uint8) {
	p.emitBytes(OP_DEFINE_GLOBAL, global)
}

func (p *Parser) namedVariable(name Token, canAssign bool) {
	arg := p.identifierConstant(name)
	if canAssign && p.match(TOKEN_EQUAL) {
		p.expression()
		p.emitBytes(OP_SET_GLOBAL, arg)
	} else {
		p.emitBytes(OP_GET_GLOBAL, arg)
	}
}

func (p *Parser) emitConstant(value Value) {
	p.emitBytes(OP_CONSTANT, p.makeConstant(value))
}

func (p *Parser) makeConstant(value Value) uint8 {
	constidx := p.compilingChunk.addConstant(value)
	if constidx > 255 {
		p.error("Too many constants in one chunk")
		return 0
	}
	return constidx
}

func (p *Parser) emitReturn() {
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
	p.emitConstant(NumberValue{value: val})

}

func loxstring(p *Parser, canAssign bool) {

	str := p.previous.lexeme()
	strobj := MakeStringObject(strings.Replace(str, "\"", "", -1))
	p.emitConstant(MakeObjectValue(&strobj))

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
