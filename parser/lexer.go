package parser

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"eml-parser/token"
)

type lexer struct {
	input  string
	offset int
	line   int
	column int
}

func newLexer(input string) *lexer {
	return &lexer{
		input:  input,
		line:   1,
		column: 1,
	}
}

func (l *lexer) next() token.Token {
	l.skipWhitespace()

	start := l.position()
	r, width := l.peek()
	if width == 0 {
		return token.Token{
			Kind: token.EOF,
			Span: token.Span{Start: start, End: start},
		}
	}

	switch {
	case isIdentifierStart(r):
		lexeme := l.scanIdentifier()
		kind := token.Ident
		if lexeme == "eml" {
			kind = token.EML
		}
		return token.Token{
			Kind:   kind,
			Lexeme: lexeme,
			Span:   token.Span{Start: start, End: l.position()},
		}
	case r == '1':
		l.advance(width)
		return token.Token{
			Kind:   token.One,
			Lexeme: "1",
			Span:   token.Span{Start: start, End: l.position()},
		}
	case r == '(':
		l.advance(width)
		return token.Token{
			Kind:   token.LParen,
			Lexeme: "(",
			Span:   token.Span{Start: start, End: l.position()},
		}
	case r == ')':
		l.advance(width)
		return token.Token{
			Kind:   token.RParen,
			Lexeme: ")",
			Span:   token.Span{Start: start, End: l.position()},
		}
	case r == ',':
		l.advance(width)
		return token.Token{
			Kind:   token.Comma,
			Lexeme: ",",
			Span:   token.Span{Start: start, End: l.position()},
		}
	case unicode.IsDigit(r):
		lexeme := l.scanDigits()
		return token.Token{
			Kind:   token.Illegal,
			Lexeme: lexeme,
			Span:   token.Span{Start: start, End: l.position()},
		}
	default:
		l.advance(width)
		return token.Token{
			Kind:   token.Illegal,
			Lexeme: string(r),
			Span:   token.Span{Start: start, End: l.position()},
		}
	}
}

func (l *lexer) skipWhitespace() {
	for {
		r, width := l.peek()
		if width == 0 || !unicode.IsSpace(r) {
			return
		}
		l.advance(width)
	}
}

func (l *lexer) scanIdentifier() string {
	start := l.offset
	for {
		r, width := l.peek()
		if width == 0 || !isIdentifierPart(r) {
			return l.input[start:l.offset]
		}
		l.advance(width)
	}
}

func (l *lexer) scanDigits() string {
	start := l.offset
	for {
		r, width := l.peek()
		if width == 0 || !unicode.IsDigit(r) {
			return l.input[start:l.offset]
		}
		l.advance(width)
	}
}

func (l *lexer) peek() (rune, int) {
	if l.offset >= len(l.input) {
		return 0, 0
	}
	r, width := utf8.DecodeRuneInString(l.input[l.offset:])
	return r, width
}

func (l *lexer) advance(width int) {
	if width <= 0 {
		return
	}

	r, _ := utf8.DecodeRuneInString(l.input[l.offset:])
	l.offset += width
	if r == '\n' {
		l.line++
		l.column = 1
		return
	}
	l.column++
}

func (l *lexer) position() token.Position {
	return token.Position{
		Offset: l.offset,
		Line:   l.line,
		Column: l.column,
	}
}

func isIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isIdentifierPart(r rune) bool {
	return isIdentifierStart(r) || unicode.IsDigit(r)
}

func illegalTokenMessage(tok token.Token) string {
	if tok.Lexeme == "" {
		return "illegal token"
	}
	if tok.Lexeme == "0" || tok.Lexeme == "2" || tok.Lexeme == "3" {
		return fmt.Sprintf("numeric literal %q is not part of the minimal paper-grounded grammar; only the distinguished constant 1 is accepted", tok.Lexeme)
	}
	return fmt.Sprintf("unexpected token %q", tok.Lexeme)
}
