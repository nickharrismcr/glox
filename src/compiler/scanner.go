package compiler

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
	TOKEN_STATIC
	TOKEN_FROM
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
	"static":     TOKEN_STATIC,
	"from":       TOKEN_FROM,
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
	TOKEN_STATIC:        "TOKEN_STATIC",
	TOKEN_FROM:          "TOKEN_FROM",
}

type Scanner struct {
	Source               string
	Start, Current, Line int
	Tokens               TokenList
	TokenIdx             int
}

type Token struct {
	Tokentype           TokenType
	Source              *string
	Start, Length, Line int
}

type TokenList struct {
	Tokens []Token
}

func MakeTokenList() TokenList {
	return TokenList{
		Tokens: []Token{},
	}
}

func (t *TokenList) Add(token Token) {
	t.Tokens = append(t.Tokens, token)
}

func (t TokenList) Size() int {
	return len(t.Tokens)
}

func (t TokenList) At(idx int) Token {
	return t.Tokens[idx]
}

func (t TokenList) Get(offset int) Token {
	return t.Tokens[t.Size()+offset]
}

func (t TokenList) Print() {

	for i, token := range t.Tokens {
		fmt.Printf("%d: %s (%s)  [%d:%d]\n", i, repr[token.Tokentype], token.Lexeme(), token.Line, token.Start)
	}
}

func (t Token) Lexeme() string {

	rv := (*t.Source)[t.Start : t.Start+t.Length]
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
		Source:   source,
		Line:     1,
		Tokens:   MakeTokenList(),
		TokenIdx: 0,
	}
	for {
		t := s.ScanToken()
		s.Tokens.Add(t)
		if t.Tokentype == TOKEN_EOF {
			break
		}
	}
	return s
}
func (s *Scanner) NextToken() Token {

	token := s.Tokens.At(s.TokenIdx)
	s.TokenIdx++
	return token
}

func (s *Scanner) ScanToken() Token {

	for {

		s.SkipWhiteSpace()
		s.Start = s.Current
		if s.IsAtEnd() {
			return s.MakeToken(TOKEN_EOF)
		}
		c := s.Advance()
		if s.IsAlpha(c) {
			return s.Identifier()
		}
		if s.IsDigit(c) {
			return s.Number()
		}
		switch c {
		case "\n":
			if !s.SkipEOL() {
				rv := s.MakeToken(TOKEN_EOL)
				s.Line++
				return rv
			}
			s.Line++
		case "(":
			return s.MakeToken(TOKEN_LEFT_PAREN)
		case ")":
			return s.MakeToken(TOKEN_RIGHT_PAREN)
		case "{":
			return s.MakeToken(TOKEN_LEFT_BRACE)
		case "}":
			return s.MakeToken(TOKEN_RIGHT_BRACE)
		case "[":
			return s.MakeToken(TOKEN_LEFT_BRACKET)
		case "]":
			return s.MakeToken(TOKEN_RIGHT_BRACKET)
		case ";":
			return s.MakeToken(TOKEN_SEMICOLON)
		case ":":
			return s.MakeToken(TOKEN_COLON)
		case ",":
			return s.MakeToken(TOKEN_COMMA)
		case ".":
			return s.MakeToken(TOKEN_DOT)
		case "-":
			return s.MakeToken(TOKEN_MINUS)
		case "+":
			return s.MakeToken(TOKEN_PLUS)
		case "%":
			return s.MakeToken(TOKEN_PERCENT)
		case "/":
			return s.MakeToken(TOKEN_SLASH)
		case "*":
			return s.MakeToken(TOKEN_STAR)
		case "!":
			if s.Match("=") {
				return s.MakeToken(TOKEN_BANG_EQUAL)
			}
			return s.MakeToken(TOKEN_BANG)
		case "=":
			if s.Match("=") {
				return s.MakeToken(TOKEN_EQUAL_EQUAL)
			}
			return s.MakeToken(TOKEN_EQUAL)
		case "<":
			if s.Match("=") {
				return s.MakeToken(TOKEN_LESS_EQUAL)
			}
			return s.MakeToken(TOKEN_LESS)
		case ">":
			if s.Match("=") {
				return s.MakeToken(TOKEN_GREATER_EQUAL)
			}
			return s.MakeToken(TOKEN_GREATER)
		case "\"":
			return s.string("\"")
		case "'":
			return s.string("'")
		default:
			return s.ErrorToken(fmt.Sprintf("Unexpected character [%s]", c))
		}
	}

}

func (s *Scanner) EOL() Token {

	s.Line++
	return s.MakeToken(TOKEN_EOL)
}

func (s *Scanner) IsAtEnd() bool {

	return s.Current == len(s.Source)
}

func (s *Scanner) MakeToken(tokentype TokenType) Token {

	rv := Token{
		Tokentype: tokentype,
		Start:     s.Start,
		Length:    s.Current - s.Start,
		Line:      s.Line,
		Source:    &s.Source,
	}
	return rv
}

func (s *Scanner) ErrorToken(message string) Token {

	return Token{
		Tokentype: TOKEN_ERROR,
		Start:     0,
		Length:    len(message),
		Line:      s.Line,
		Source:    &message,
	}
}

func (s *Scanner) Advance() string {

	s.Current++
	return s.Source[s.Current-1 : s.Current]
}

func (s *Scanner) Match(expected string) bool {

	if s.IsAtEnd() {
		return false
	}
	if s.Source[s.Current:s.Current+1] != expected {
		return false
	}
	s.Current++
	return true
}

func (s *Scanner) SkipWhiteSpace() {

	for {
		c := s.Peek()
		switch c {
		case " ":
			s.Advance()
		case "\r":
			s.Advance()
		case "\t":
			s.Advance()
		case "/":
			if s.PeekNext() == "/" {
				for s.Peek() != "\n" && !s.IsAtEnd() {
					s.Advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}

}

func (s *Scanner) SkipEOL() bool {

	if s.Tokens.Size() < 2 {
		return true
	}
	prev := s.Tokens.Get(-1).Tokentype

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

func (s *Scanner) Peek() string {

	if s.Current == len(s.Source) {
		return "\\0"
	}
	return s.Source[s.Current : s.Current+1]
}

func (s *Scanner) PeekNext() string {

	if s.Current == len(s.Source) {
		return "\\0"
	}
	return s.Source[s.Current+1 : s.Current+2]
}

func (s *Scanner) string(which string) Token {

	for s.Peek() != which && !s.IsAtEnd() {
		if s.Peek() == "\n" {
			s.Line++
		}
		s.Advance()
	}
	if s.IsAtEnd() {
		return s.ErrorToken("Unterminated string")
	}
	s.Advance()
	return s.MakeToken(TOKEN_STRING)
}

func (s *Scanner) IsDigit(c string) bool {

	return (c >= "0") && (c <= "9")
}

func (s *Scanner) Number() Token {

	for s.IsDigit(s.Peek()) {
		s.Advance()
	}
	if s.Peek() == "." && s.IsDigit(s.PeekNext()) {
		s.Advance()
	} else {
		return s.MakeToken(TOKEN_INT)
	}
	for s.IsDigit(s.Peek()) {
		s.Advance()
	}
	return s.MakeToken(TOKEN_FLOAT)
}

func (s *Scanner) IsAlpha(c string) bool {

	return (c >= "a" && c <= "z") ||
		(c >= "A" && c <= "Z") ||
		(c == "_")
}

func (s *Scanner) Identifier() Token {

	for s.IsAlpha(s.Peek()) || s.IsDigit(s.Peek()) {
		s.Advance()
	}
	return s.MakeToken(s.IdentifierType())
}

func (s *Scanner) IdentifierType() TokenType {

	id := s.Source[s.Start:s.Current]
	if tokentype, ok := keywords[id]; ok {
		return tokentype
	}
	return TOKEN_IDENTIFIER
}

func SyntheticToken(src string) Token {

	return Token{
		Tokentype: TOKEN_THIS,
		Source:    &src,
		Start:     0,
		Length:    4,
		Line:      0,
	}
}

func PrintTokens(source string) {

	s := NewScanner(source)
	for {
		t := s.ScanToken()
		if t.Tokentype == TOKEN_EOF {
			break
		}
	}
	s.Tokens.Print()

}
