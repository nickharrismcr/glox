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
	TOKEN_PLUS_EQUAL
	TOKEN_MINUS_EQUAL
	TOKEN_STAR_EQUAL
	TOKEN_SLASH_EQUAL
	TOKEN_PERCENT_EQUAL
	TOKEN_SEMICOLON
	TOKEN_SLASH
	TOKEN_STAR
	TOKEN_COLON
	TOKEN_QUESTION
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
	TOKEN_PLUS_PLUS // ++
	TOKEN_AMPERSAND // &
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
	TOKEN_PLUS_EQUAL:    "TOKEN_PLUS_EQUAL",
	TOKEN_MINUS_EQUAL:   "TOKEN_MINUS_EQUAL",
	TOKEN_STAR_EQUAL:    "TOKEN_STAR_EQUAL",
	TOKEN_SLASH_EQUAL:   "TOKEN_SLASH_EQUAL",
	TOKEN_PERCENT_EQUAL: "TOKEN_PERCENT_EQUAL",
	TOKEN_SEMICOLON:     "TOKEN_SEMICOLON",
	TOKEN_SLASH:         "TOKEN_SLASH",
	TOKEN_STAR:          "TOKEN_STAR",
	TOKEN_COLON:         "TOKEN_COLON",
	TOKEN_QUESTION:      "TOKEN_QUESTION",
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
	TOKEN_PLUS_PLUS:     "TOKEN_PLUS_PLUS",
	TOKEN_AMPERSAND:     "TOKEN_AMPERSAND",
}

type Scanner struct {
	Source               string
	Start, Current, Line int
	Tokens               TokenList
	TokenIdx             int
	pending              []Token // queued tokens from string interpolation desugaring
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

	// The final entry in Tokens is always TOKEN_EOF (see NewScanner). Once
	// reached, keep returning it instead of advancing past the end of the
	// slice -- callers such as parsePrecedence() call advance() unconditionally,
	// so TokenIdx can otherwise be driven past the last valid index.
	if s.TokenIdx >= s.Tokens.Size()-1 {
		return s.Tokens.At(s.Tokens.Size() - 1)
	}
	token := s.Tokens.At(s.TokenIdx)
	s.TokenIdx++
	return token
}

// SkipToEnd abandons the remaining token stream and jumps straight to the
// trailing TOKEN_EOF. Used by the parser to bail out of a malformed parse
// (e.g. runaway expression nesting) without consuming the remaining tokens
// one at a time or risking further recursion into them.
func (s *Scanner) SkipToEnd() Token {

	s.TokenIdx = s.Tokens.Size() - 1
	return s.Tokens.At(s.Tokens.Size() - 1)
}

func (s *Scanner) ScanToken() Token {

	// Drain any tokens queued by string interpolation desugaring first.
	if len(s.pending) > 0 {
		t := s.pending[0]
		s.pending = s.pending[1:]
		return t
	}

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
		case "?":
			return s.MakeToken(TOKEN_QUESTION)
		case ",":
			return s.MakeToken(TOKEN_COMMA)
		case ".":
			return s.MakeToken(TOKEN_DOT)
		case "-":
			if s.Match("=") {
				return s.MakeToken(TOKEN_MINUS_EQUAL)
			}
			return s.MakeToken(TOKEN_MINUS)
		case "+":
			if s.Match("=") {
				return s.MakeToken(TOKEN_PLUS_EQUAL)
			}
			if s.Match("+") {
				return s.MakeToken(TOKEN_PLUS_PLUS)
			}
			return s.MakeToken(TOKEN_PLUS)
		case "&":
			return s.MakeToken(TOKEN_AMPERSAND)
		case "%":
			if s.Match("=") {
				return s.MakeToken(TOKEN_PERCENT_EQUAL)
			}
			return s.MakeToken(TOKEN_PERCENT)
		case "/":
			if s.Match("=") {
				return s.MakeToken(TOKEN_SLASH_EQUAL)
			}
			return s.MakeToken(TOKEN_SLASH)
		case "*":
			if s.Match("=") {
				return s.MakeToken(TOKEN_STAR_EQUAL)
			}
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
		prev == TOKEN_QUESTION ||
		prev == TOKEN_EQUAL ||
		prev == TOKEN_MINUS ||
		prev == TOKEN_PLUS ||
		prev == TOKEN_SLASH ||
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

// interpSegment is one piece of an interpolated string literal: either a run of
// literal text (isExpr == false) or an embedded ${ ... } expression already
// tokenised into toks (isExpr == true).
type interpSegment struct {
	isExpr bool
	text   string
	toks   []Token
}

func (s *Scanner) string(which string) Token {

	// Opening quote has already been consumed by ScanToken. Walk the literal,
	// splitting on ${ ... } interpolations and collapsing the $$ escape.
	var lit strings.Builder
	var segments []interpSegment
	hasInterp := false
	hadEscape := false

	for {
		if s.IsAtEnd() {
			return s.ErrorToken("Unterminated string")
		}
		c := s.Peek()
		if c == which {
			s.Advance() // consume closing quote
			break
		}
		if c == "$" && s.PeekNext() == "$" {
			// $$ escapes to a single literal $
			hadEscape = true
			s.Advance()
			s.Advance()
			lit.WriteString("$")
			continue
		}
		if c == "$" && s.PeekNext() == "{" {
			hasInterp = true
			segments = append(segments, interpSegment{isExpr: false, text: lit.String()})
			lit.Reset()
			s.Advance() // $
			s.Advance() // {
			exprSrc, ok := s.scanInterpExpr()
			if !ok {
				return s.ErrorToken("Unterminated interpolation")
			}
			if strings.TrimSpace(exprSrc) == "" {
				return s.ErrorToken("Empty interpolation expression")
			}
			segments = append(segments, interpSegment{isExpr: true, toks: s.scanInterpTokens(exprSrc)})
			continue
		}
		if c == "\n" {
			s.Line++
		}
		lit.WriteString(c)
		s.Advance()
	}

	if !hasInterp {
		if !hadEscape {
			// Fast path: no interpolation and no escapes — emit the literal
			// exactly as before, spanning the original source (quotes included).
			return s.MakeToken(TOKEN_STRING)
		}
		// Only $$ escapes: emit a single synthetic string with collapsed content.
		return s.synthString(lit.String(), which)
	}

	// Flush the trailing literal run.
	segments = append(segments, interpSegment{isExpr: false, text: lit.String()})

	// Synthesise: ( seg0 & seg1 & ... ) where literal segments become string
	// constants and expression segments become str( <expr tokens> ).
	out := []Token{s.synthToken(TOKEN_LEFT_PAREN, "(")}
	first := true
	amp := func() {
		if !first {
			out = append(out, s.synthToken(TOKEN_AMPERSAND, "&"))
		}
		first = false
	}
	for _, seg := range segments {
		if seg.isExpr {
			amp()
			out = append(out, s.synthToken(TOKEN_STR, "str"), s.synthToken(TOKEN_LEFT_PAREN, "("))
			out = append(out, seg.toks...)
			out = append(out, s.synthToken(TOKEN_RIGHT_PAREN, ")"))
		} else if seg.text != "" {
			amp()
			out = append(out, s.synthString(seg.text, which))
		}
	}
	out = append(out, s.synthToken(TOKEN_RIGHT_PAREN, ")"))

	// Return the first token; queue the rest for subsequent ScanToken calls.
	s.pending = append(s.pending, out[1:]...)
	return out[0]
}

// scanInterpExpr consumes source from just after "${" through the matching "}"
// (tracking brace depth and skipping nested string literals so their braces and
// the outer quote do not interfere). It returns the expression source between
// the braces, or ok == false if the string/EOF ends first.
func (s *Scanner) scanInterpExpr() (string, bool) {

	start := s.Current
	depth := 1
	for !s.IsAtEnd() {
		c := s.Peek()
		if c == "\"" || c == "'" {
			s.Advance() // opening quote
			for !s.IsAtEnd() && s.Peek() != c {
				if s.Peek() == "\n" {
					s.Line++
				}
				s.Advance()
			}
			if s.IsAtEnd() {
				return "", false
			}
			s.Advance() // closing quote
			continue
		}
		if c == "{" {
			depth++
			s.Advance()
			continue
		}
		if c == "}" {
			depth--
			if depth == 0 {
				expr := s.Source[start:s.Current]
				s.Advance() // consume closing }
				return expr, true
			}
			s.Advance()
			continue
		}
		if c == "\n" {
			s.Line++
		}
		s.Advance()
	}
	return "", false
}

// scanInterpTokens tokenises an interpolated expression's source by running a
// fresh scanner over it and stripping the trailing EOL/EOF tokens. Nested
// interpolation is handled recursively. All tokens are re-lined to the current
// source line for sensible error reporting.
func (s *Scanner) scanInterpTokens(exprSrc string) []Token {

	sub := NewScanner(exprSrc)
	toks := sub.Tokens.Tokens
	end := len(toks)
	for end > 0 && (toks[end-1].Tokentype == TOKEN_EOF || toks[end-1].Tokentype == TOKEN_EOL) {
		end--
	}
	toks = toks[:end]
	for i := range toks {
		toks[i].Line = s.Line
	}
	return toks
}

// synthToken builds a synthetic token backed by its own lexeme string.
func (s *Scanner) synthToken(tt TokenType, lexeme string) Token {

	src := lexeme
	return Token{
		Tokentype: tt,
		Source:    &src,
		Start:     0,
		Length:    len(src),
		Line:      s.Line,
	}
}

// synthString builds a synthetic TOKEN_STRING whose lexeme is content wrapped in
// the delimiter, so the compiler's loxstring (which strips the first/last char)
// recovers content. content cannot contain the delimiter, so this is unambiguous.
func (s *Scanner) synthString(content, which string) Token {

	src := which + content + which
	return Token{
		Tokentype: TOKEN_STRING,
		Source:    &src,
		Start:     0,
		Length:    len(src),
		Line:      s.Line,
	}
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
