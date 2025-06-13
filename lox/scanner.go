package lox

import (
	"fmt"
	"strings"
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
	TOKEN_EOL
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
	TOKEN_IMPORT
	TOKEN_TRY
	TOKEN_EXCEPT
	TOKEN_AS
	TOKEN_FINALLY
	TOKEN_RAISE
	TOKEN_FOREACH
	TOKEN_IN
	TOKEN_BREAKPOINT
)

var keywords = map[string]TokenType{
	"and":        TOKEN_AND,
	"class":      TOKEN_CLASS,
	"else":       TOKEN_ELSE,
	"if":         TOKEN_IF,
	"nil":        TOKEN_NIL,
	"or":         TOKEN_OR,
	"print":      TOKEN_PRINT,
	"return":     TOKEN_RETURN,
	"super":      TOKEN_SUPER,
	"var":        TOKEN_VAR,
	"while":      TOKEN_WHILE,
	"false":      TOKEN_FALSE,
	"for":        TOKEN_FOR,
	"fun":        TOKEN_FUNC,
	"func":       TOKEN_FUNC,
	"this":       TOKEN_THIS,
	"true":       TOKEN_TRUE,
	"const":      TOKEN_CONST,
	"break":      TOKEN_BREAK,
	"continue":   TOKEN_CONTINUE,
	"str":        TOKEN_STR,
	"import":     TOKEN_IMPORT,
	"try":        TOKEN_TRY,
	"except":     TOKEN_EXCEPT,
	"finally":    TOKEN_FINALLY,
	"raise":      TOKEN_RAISE,
	"as":         TOKEN_AS,
	"foreach":    TOKEN_FOREACH,
	"in":         TOKEN_IN,
	"breakpoint": TOKEN_BREAKPOINT,
}

var repr = map[TokenType]string{
	TOKEN_LEFT_PAREN:    "TOKEN_LEFT_PAREN ",
	TOKEN_RIGHT_PAREN:   "TOKEN_RIGHT_PAREN",
	TOKEN_LEFT_BRACE:    "TOKEN_LEFT_BRACE",
	TOKEN_RIGHT_BRACE:   "TOKEN_RIGHT_BRACE",
	TOKEN_LEFT_BRACKET:  "TOKEN_LEFT_BRACKET",
	TOKEN_RIGHT_BRACKET: "TOKEN_RIGHT_BRACKET",
	TOKEN_COMMA:         "TOKEN_COMMA",
	TOKEN_DOT:           "TOKEN_DOT",
	TOKEN_PERCENT:       "TOKEN_PERCENT",
	TOKEN_MINUS:         "TOKEN_MINUS",
	TOKEN_PLUS:          "TOKEN_PLUS",
	TOKEN_SEMICOLON:     "TOKEN_SEMICOLON",
	TOKEN_SLASH:         "TOKEN_SLASH",
	TOKEN_STAR:          "TOKEN_STAR",
	TOKEN_COLON:         "TOKEN_COLON",
	TOKEN_EOL:           "TOKEN_EOL",
	TOKEN_BANG:          "TOKEN_BANG",
	TOKEN_BANG_EQUAL:    "TOKEN_BANG_EQUAL",
	TOKEN_EQUAL:         "TOKEN_EQUAL",
	TOKEN_EQUAL_EQUAL:   "TOKEN_EQUAL_EQUAL",
	TOKEN_GREATER:       "TOKEN_GREATER",
	TOKEN_GREATER_EQUAL: "TOKEN_GREATER_EQUAL",
	TOKEN_LESS:          "TOKEN_LESS",
	TOKEN_LESS_EQUAL:    "TOKEN_LESS_EQUAL",
	TOKEN_IDENTIFIER:    "TOKEN_IDENTIFIER",
	TOKEN_STRING:        "TOKEN_STRING",
	TOKEN_INT:           "TOKEN_INT",
	TOKEN_FLOAT:         "TOKEN_FLOAT",
	TOKEN_AND:           "TOKEN_AND",
	TOKEN_CLASS:         "TOKEN_CLASS",
	TOKEN_ELSE:          "TOKEN_ELSE",
	TOKEN_FALSE:         "TOKEN_FALSE",
	TOKEN_FOR:           "TOKEN_FOR",
	TOKEN_FUNC:          "TOKEN_FUNC",
	TOKEN_IF:            "TOKEN_IF",
	TOKEN_NIL:           "TOKEN_NIL",
	TOKEN_OR:            "TOKEN_OR",
	TOKEN_PRINT:         "TOKEN_PRINT",
	TOKEN_RETURN:        "TOKEN_RETURN",
	TOKEN_SUPER:         "TOKEN_SUPER",
	TOKEN_THIS:          "TOKEN_THIS",
	TOKEN_TRUE:          "TOKEN_TRUE",
	TOKEN_VAR:           "TOKEN_VAR",
	TOKEN_WHILE:         "TOKEN_WHILE",
	TOKEN_ERROR:         "TOKEN_ERROR",
	TOKEN_EOF:           "TOKEN_EOF",
	TOKEN_CONST:         "TOKEN_CONST",
	TOKEN_BREAK:         "TOKEN_BREAK",
	TOKEN_CONTINUE:      "TOKEN_CONTINUE",
	TOKEN_STR:           "TOKEN_STR",
	TOKEN_IMPORT:        "TOKEN_IMPORT",
	TOKEN_TRY:           "TOKEN_TRY",
	TOKEN_EXCEPT:        "TOKEN_EXCEPT",
	TOKEN_AS:            "TOKEN_AS",
	TOKEN_FINALLY:       "TOKEN_FINALLY",
	TOKEN_RAISE:         "TOKEN_RAISE",
	TOKEN_FOREACH:       "TOKEN_FOREACH",
	TOKEN_IN:            "TOKEN_IN",
	TOKEN_BREAKPOINT:    "TOKEN_BREAKPOINT",
}

type Scanner struct {
	source               string
	start, current, line int
	tokens               TokenList
	tokenIdx             int
}

type Token struct {
	tokentype           TokenType
	source              *string
	start, length, line int
}

type TokenList struct {
	tokens []Token
}

func MakeTokenList() TokenList {
	return TokenList{
		tokens: []Token{},
	}
}

func (t *TokenList) add(token Token) {
	t.tokens = append(t.tokens, token)
}

func (t TokenList) size() int {
	return len(t.tokens)
}

func (t TokenList) at(idx int) Token {
	return t.tokens[idx]
}

func (t TokenList) get(offset int) Token {
	return t.tokens[t.size()+offset]
}

func (t TokenList) print() {

	for i, token := range t.tokens {
		fmt.Printf("%d: %s (%s)  [%d:%d]\n", i, repr[token.tokentype], token.lexeme(), token.line, token.start)
	}
}

func (t Token) lexeme() string {

	rv := (*t.source)[t.start : t.start+t.length]
	if rv == "\n" {
		return "\\n"
	}
	return rv
}

func NewScanner(source string) *Scanner {

	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	source = source + "\n"
	s := &Scanner{
		source:   source,
		line:     1,
		tokens:   MakeTokenList(),
		tokenIdx: 0,
	}
	for {
		t := s.scanToken()
		s.tokens.add(t)
		if t.tokentype == TOKEN_EOF {
			break
		}
	}
	return s
}
func (s *Scanner) nextToken() Token {

	token := s.tokens.at(s.tokenIdx)
	s.tokenIdx++
	return token
}

func (s *Scanner) scanToken() Token {

	for {

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
		case "\n":
			if !s.skipEOL() {
				rv := s.makeToken(TOKEN_EOL)
				s.line++
				return rv
			}
			s.line++
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
		default:
			return s.errorToken(fmt.Sprintf("Unexpected character [%s]", c))
		}
	}

}

func (s *Scanner) EOL() Token {

	s.line++
	return s.makeToken(TOKEN_EOL)
}

func (s *Scanner) isAtEnd() bool {

	return s.current == len(s.source)
}

func (s *Scanner) makeToken(tokentype TokenType) Token {

	rv := Token{
		tokentype: tokentype,
		start:     s.start,
		length:    s.current - s.start,
		line:      s.line,
		source:    &s.source,
	}
	return rv
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

func (s *Scanner) skipEOL() bool {

	if s.tokens.size() < 2 {
		return true
	}
	prev := s.tokens.get(-1).tokentype

	if prev == TOKEN_LEFT_BRACE ||
		prev == TOKEN_LEFT_PAREN ||
		prev == TOKEN_LEFT_BRACKET ||
		prev == TOKEN_COMMA ||
		prev == TOKEN_SEMICOLON ||
		prev == TOKEN_COLON ||
		prev == TOKEN_EOL ||
		prev == TOKEN_EQUAL ||
		prev == TOKEN_MINUS ||
		prev == TOKEN_PLUS ||
		prev == TOKEN_SLASH ||
		prev == TOKEN_STAR ||
		prev == TOKEN_PERCENT {
		return true
	}
	return false
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

func PrintTokens(source string) {

	s := NewScanner(source)
	for {
		t := s.scanToken()
		if t.tokentype == TOKEN_EOF {
			break
		}
	}
	s.tokens.print()

}
