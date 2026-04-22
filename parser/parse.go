package parser

import (
	"fmt"

	"eml-parser/ast"
	"eml-parser/token"
)

//go:generate go run golang.org/x/tools/cmd/goyacc -o eml_parser.go -p eml eml.y

// Error is a parse or lex error with source position.
type Error struct {
	Message  string
	Position token.Position
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s at line %d, column %d", e.Message, e.Position.Line, e.Position.Column)
}

// ParseString parses the minimal EML language into an AST.
func ParseString(input string) (ast.Expr, error) {
	driver := &parserDriver{lexer: newLexer(input)}
	if emlParse(driver) != 0 {
		if driver.err != nil {
			return nil, driver.err
		}
		return nil, &Error{
			Message:  "parse failed",
			Position: driver.last.Span.Start,
		}
	}
	if driver.err != nil {
		return nil, driver.err
	}
	if driver.result == nil {
		return nil, &Error{
			Message:  "parser produced no expression",
			Position: token.Position{Line: 1, Column: 1},
		}
	}
	return driver.result, nil
}

type parserDriver struct {
	lexer  *lexer
	result ast.Expr
	err    error
	last   token.Token
}

func (d *parserDriver) Lex(lval *emlSymType) int {
	tok := d.lexer.next()
	d.last = tok
	lval.tok = tok

	switch tok.Kind {
	case token.Illegal:
		d.setError(&Error{
			Message:  illegalTokenMessage(tok),
			Position: tok.Span.Start,
		})
		return ILLEGAL
	case token.EOF:
		return 0
	case token.Ident:
		return IDENT
	case token.One:
		return ONE
	case token.EML:
		return EML
	case token.LParen:
		return LPAREN
	case token.RParen:
		return RPAREN
	case token.Comma:
		return COMMA
	default:
		d.setError(&Error{
			Message:  fmt.Sprintf("unsupported token kind %s", tok.Kind),
			Position: tok.Span.Start,
		})
		return ILLEGAL
	}
}

func (d *parserDriver) Error(msg string) {
	if d.err != nil {
		return
	}
	pos := d.last.Span.Start
	if pos.Line == 0 {
		pos = token.Position{Line: 1, Column: 1}
	}
	d.err = &Error{
		Message:  msg,
		Position: pos,
	}
}

func (d *parserDriver) setError(err error) {
	if d.err == nil {
		d.err = err
	}
}

func toASTSpan(span token.Span) ast.Span {
	return ast.Span{
		Start: toASTPosition(span.Start),
		End:   toASTPosition(span.End),
	}
}

func toASTPosition(pos token.Position) ast.Position {
	return ast.Position{
		Offset: pos.Offset,
		Line:   pos.Line,
		Column: pos.Column,
	}
}
