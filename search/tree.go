package search

import "eml-parser/ast"

// Stats summarizes the structure of a raw EML tree.
type Stats struct {
	NodeCount int
	TreeDepth int
	LeafCount int
}

// TreeStats computes structural metrics over a raw EML AST.
func TreeStats(expr ast.Expr) Stats {
	nodes, depth, leaves := treeStats(expr)
	return Stats{
		NodeCount: nodes,
		TreeDepth: depth,
		LeafCount: leaves,
	}
}

func treeStats(expr ast.Expr) (nodes, depth, leaves int) {
	switch n := expr.(type) {
	case ast.One, ast.Variable:
		return 1, 1, 1
	case ast.Apply:
		leftNodes, leftDepth, leftLeaves := treeStats(n.Left)
		rightNodes, rightDepth, rightLeaves := treeStats(n.Right)
		return 1 + leftNodes + rightNodes, 1 + max(leftDepth, rightDepth), leftLeaves + rightLeaves
	default:
		return 0, 0, 0
	}
}

// CanonicalKey returns a stable string key for a raw EML AST.
func CanonicalKey(expr ast.Expr) string {
	return expr.String()
}

// Equal reports structural equality between two raw EML trees.
func Equal(a, b ast.Expr) bool {
	switch av := a.(type) {
	case ast.One:
		_, ok := b.(ast.One)
		return ok
	case ast.Variable:
		bv, ok := b.(ast.Variable)
		return ok && av.Name == bv.Name
	case ast.Apply:
		bv, ok := b.(ast.Apply)
		return ok && Equal(av.Left, bv.Left) && Equal(av.Right, bv.Right)
	default:
		return false
	}
}

// Subtrees returns every subtree in pre-order traversal.
func Subtrees(expr ast.Expr) []ast.Expr {
	var out []ast.Expr
	walk(expr, func(node ast.Expr) {
		out = append(out, clone(exprOf(node)))
	})
	return out
}

func walk(expr ast.Expr, visit func(ast.Expr)) {
	visit(expr)
	if app, ok := expr.(ast.Apply); ok {
		walk(app.Left, visit)
		walk(app.Right, visit)
	}
}

func exprOf(expr ast.Expr) ast.Expr {
	return expr
}

func clone(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case ast.One:
		return ast.One{}
	case ast.Variable:
		return ast.Variable{Name: n.Name}
	case ast.Apply:
		return ast.Apply{
			Left:  clone(n.Left),
			Right: clone(n.Right),
		}
	default:
		return nil
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
