package lox

import (
	"fmt"
)

type TokenType int

const (
	// Single character tokens
	TOKEN_LEFT_PAREN TokenType = iota
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_LEFT_BRACKET
	TOKEN_RIGHT_BRACKET
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_PERCENT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_SEMICOLON
	TOKEN_SLASH
	TOKEN_STAR
	TOKEN_COLON
	// One or two character tokens.
	TOKEN_BANG
	TOKEN_BANG_EQUAL
	TOKEN_EQUAL
	TOKEN_EQUAL_EQUAL
	TOKEN_GREATER
	TOKEN_GREATER_EQUAL
	TOKEN_LESS
	TOKEN_LESS_EQUAL
	// Literals.
	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_INT
	TOKEN_FLOAT
	// Keywords.
	TOKEN_AND
	TOKEN_CLASS
	TOKEN_ELSE
	TOKEN_FALSE
	TOKEN_FOR
	TOKEN_FUNC
	TOKEN_IF
	TOKEN_NIL
	TOKEN_OR
	TOKEN_PRINT
	TOKEN_RETURN
	TOKEN_SUPER
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_VAR
	TOKEN_WHILE
	TOKEN_ERROR
	TOKEN_EOF
	TOKEN_CONST
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_STR
)

var keywords = map[string]TokenType{
	"and":      TOKEN_AND,
	"class":    TOKEN_CLASS,
	"else":     TOKEN_ELSE,
	"if":       TOKEN_IF,
	"nil":      TOKEN_NIL,
	"or":       TOKEN_OR,
	"print":    TOKEN_PRINT,
	"return":   TOKEN_RETURN,
	"super":    TOKEN_SUPER,
	"var":      TOKEN_VAR,
	"while":    TOKEN_WHILE,
	"false":    TOKEN_FALSE,
	"for":      TOKEN_FOR,
	"func":     TOKEN_FUNC,
	"this":     TOKEN_THIS,
	"true":     TOKEN_TRUE,
	"const":    TOKEN_CONST,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"str":      TOKEN_STR,
}

type Scanner struct {
	source               string
	start, current, line int
}

type Token struct {
	tokentype           TokenType
	source              *string
	start, length, line int
}

func (t Token) lexeme() string {

	return (*t.source)[t.start : t.start+t.length]
}

func NewScanner(source string) *Scanner {

	return &Scanner{
		source: source,
		line:   1,
	}
}

func (s *Scanner) scanToken() Token {

	s.skipWhiteSpace()
	s.start = s.current
	if s.isAtEnd() {
		return s.makeToken(TOKEN_EOF)
	}
	c := s.advance()
	if s.isAlpha(c) {
		return s.identifier()
	}
	if s.isDigit(c) {
		return s.number()
	}
	switch c {
	case "(":
		return s.makeToken(TOKEN_LEFT_PAREN)
	case ")":
		return s.makeToken(TOKEN_RIGHT_PAREN)
	case "{":
		return s.makeToken(TOKEN_LEFT_BRACE)
	case "}":
		return s.makeToken(TOKEN_RIGHT_BRACE)
	case "[":
		return s.makeToken(TOKEN_LEFT_BRACKET)
	case "]":
		return s.makeToken(TOKEN_RIGHT_BRACKET)
	case ";":
		return s.makeToken(TOKEN_SEMICOLON)
	case ":":
		return s.makeToken(TOKEN_COLON)
	case ",":
		return s.makeToken(TOKEN_COMMA)
	case ".":
		return s.makeToken(TOKEN_DOT)
	case "-":
		return s.makeToken(TOKEN_MINUS)
	case "+":
		return s.makeToken(TOKEN_PLUS)
	case "%":
		return s.makeToken(TOKEN_PERCENT)
	case "/":
		return s.makeToken(TOKEN_SLASH)
	case "*":
		return s.makeToken(TOKEN_STAR)
	case "!":
		if s.match("=") {
			return s.makeToken(TOKEN_BANG_EQUAL)
		}
		return s.makeToken(TOKEN_BANG)
	case "=":
		if s.match("=") {
			return s.makeToken(TOKEN_EQUAL_EQUAL)
		}
		return s.makeToken(TOKEN_EQUAL)
	case "<":
		if s.match("=") {
			return s.makeToken(TOKEN_LESS_EQUAL)
		}
		return s.makeToken(TOKEN_LESS)
	case ">":
		if s.match("=") {
			return s.makeToken(TOKEN_GREATER_EQUAL)
		}
		return s.makeToken(TOKEN_GREATER)
	case "\"":
		return s.string()
	}
	return s.errorToken(fmt.Sprintf("Unexpected character [%s]", c))
}

func (s *Scanner) isAtEnd() bool {

	return s.current == len(s.source)
}

func (s *Scanner) makeToken(tokentype TokenType) Token {

	return Token{
		tokentype: tokentype,
		start:     s.start,
		length:    s.current - s.start,
		line:      s.line,
		source:    &s.source,
	}
}

func (s *Scanner) errorToken(message string) Token {

	return Token{
		tokentype: TOKEN_ERROR,
		start:     0,
		length:    len(message),
		line:      s.line,
		source:    &message,
	}
}

func (s *Scanner) advance() string {

	s.current++
	return s.source[s.current-1 : s.current]
}

func (s *Scanner) match(expected string) bool {

	if s.isAtEnd() {
		return false
	}
	if s.source[s.current:s.current+1] != expected {
		return false
	}
	s.current++
	return true
}

func (s *Scanner) skipWhiteSpace() {

	for {
		c := s.peek()
		switch c {
		case " ":
			s.advance()
		case "\r":
			s.advance()
		case "\t":
			s.advance()
		case "\n":
			s.line++
			s.advance()
		case "/":
			if s.peekNext() == "/" {
				for s.peek() != "\n" && !s.isAtEnd() {
					s.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}

}

func (s *Scanner) peek() string {

	if s.current == len(s.source) {
		return "\\0"
	}
	return s.source[s.current : s.current+1]
}

func (s *Scanner) peekNext() string {

	if s.current == len(s.source) {
		return "\\0"
	}
	return s.source[s.current+1 : s.current+2]
}

func (s *Scanner) string() Token {

	for s.peek() != "\"" && !s.isAtEnd() {
		if s.peek() == "\n" {
			s.line++
		}
		s.advance()
	}
	if s.isAtEnd() {
		return s.errorToken("Unterminated string")
	}
	s.advance()
	return s.makeToken(TOKEN_STRING)
}

func (s *Scanner) isDigit(c string) bool {

	return (c >= "0") && (c <= "9")
}

func (s *Scanner) number() Token {

	for s.isDigit(s.peek()) {
		s.advance()
	}
	if s.peek() == "." && s.isDigit(s.peekNext()) {
		s.advance()
	} else {
		return s.makeToken(TOKEN_INT)
	}
	for s.isDigit(s.peek()) {
		s.advance()
	}
	return s.makeToken(TOKEN_FLOAT)
}

func (s *Scanner) isAlpha(c string) bool {

	return (c >= "a" && c <= "z") ||
		(c >= "A" && c <= "Z") ||
		(c == "_")
}

func (s *Scanner) identifier() Token {

	for s.isAlpha(s.peek()) || s.isDigit(s.peek()) {
		s.advance()
	}
	return s.makeToken(s.identifierType())
}

func (s *Scanner) identifierType() TokenType {

	id := s.source[s.start:s.current]
	if tokentype, ok := keywords[id]; ok {
		return tokentype
	}
	return TOKEN_IDENTIFIER
}

func syntheticToken(src string) Token {

	return Token{
		tokentype: TOKEN_THIS,
		source:    &src,
		start:     0,
		length:    4,
		line:      0,
	}
}
