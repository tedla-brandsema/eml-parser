package ast

import "fmt"

// Position identifies a byte offset and human-readable line/column pair.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span marks the source extent for an AST node.
type Span struct {
	Start Position
	End   Position
}

// Expr is the common interface for all EML expressions.
//
// The initial grammar is intentionally minimal and follows the paper's core
// representation: variables, the distinguished constant 1, and binary EML
// application nodes.
type Expr interface {
	expr()
	SourceSpan() Span
	String() string
}

// One is the distinguished constant terminal used by the paper's basis.
type One struct {
	Span Span
}

func (One) expr() {}

func (n One) SourceSpan() Span { return n.Span }

func (One) String() string { return "1" }

// Variable is an input terminal symbol such as x or y.
type Variable struct {
	Name string
	Span Span
}

func (Variable) expr() {}

func (n Variable) SourceSpan() Span { return n.Span }

func (n Variable) String() string { return n.Name }

// Apply is a binary EML node, corresponding to eml(left, right).
type Apply struct {
	Left  Expr
	Right Expr
	Span  Span
}

func (Apply) expr() {}

func (n Apply) SourceSpan() Span { return n.Span }

func (n Apply) String() string {
	return fmt.Sprintf("eml(%s, %s)", n.Left, n.Right)
}
