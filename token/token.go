package token

import "fmt"

// Kind enumerates lexical token classes for the minimal EML grammar.
type Kind int

const (
	Illegal Kind = iota
	EOF
	Ident
	One
	EML
	LParen
	RParen
	Comma
)

func (k Kind) String() string {
	switch k {
	case Illegal:
		return "illegal"
	case EOF:
		return "eof"
	case Ident:
		return "identifier"
	case One:
		return "one"
	case EML:
		return "eml"
	case LParen:
		return "("
	case RParen:
		return ")"
	case Comma:
		return ","
	default:
		return fmt.Sprintf("token(%d)", k)
	}
}

// Position identifies a location in the source text.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span marks the source extent of a token.
type Span struct {
	Start Position
	End   Position
}

// Token is the unit produced by the lexer.
type Token struct {
	Kind   Kind
	Lexeme string
	Span   Span
}
