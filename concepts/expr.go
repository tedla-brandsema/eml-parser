package concepts

import "fmt"

// Expr is a concept-layer expression. It may reference named concepts and is
// reduced to raw EML ASTs by a Registry.
type Expr interface {
	expr()
	String() string
}

// One is the distinguished constant terminal at the concept layer.
type One struct{}

func (One) expr() {}

func (One) String() string { return "1" }

// Param references a parameter in a concept definition.
type Param struct {
	Name string
}

func (Param) expr() {}

func (p Param) String() string { return p.Name }

// Apply is a raw EML node at the concept layer.
type Apply struct {
	Left  Expr
	Right Expr
}

func (Apply) expr() {}

func (n Apply) String() string {
	return fmt.Sprintf("eml(%s, %s)", n.Left, n.Right)
}

// Call invokes another named concept from within a concept definition.
type Call struct {
	Name string
	Args []Expr
}

func (Call) expr() {}

func (c Call) String() string {
	if len(c.Args) == 0 {
		return c.Name
	}
	return fmt.Sprintf("%s(%s)", c.Name, joinExprs(c.Args))
}

func joinExprs(exprs []Expr) string {
	if len(exprs) == 0 {
		return ""
	}
	out := exprs[0].String()
	for i := 1; i < len(exprs); i++ {
		out += ", " + exprs[i].String()
	}
	return out
}

// Convenience builders.
func P(name string) Param                     { return Param{Name: name} }
func EML(left, right Expr) Apply              { return Apply{Left: left, Right: right} }
func Ref(name string, args ...Expr) Call      { return Call{Name: name, Args: args} }
func ConstOne() One                           { return One{} }
