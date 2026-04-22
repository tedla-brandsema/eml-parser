package normalize

import "eml-parser/ast"

// Expr normalizes a raw EML AST.
func Expr(expr ast.Expr) ast.Expr {
	for {
		next, changed := rewrite(expr)
		if !changed {
			return next
		}
		expr = next
	}
}

func rewrite(expr ast.Expr) (ast.Expr, bool) {
	switch n := expr.(type) {
	case ast.One:
		return ast.One{}, false
	case ast.Variable:
		return ast.Variable{Name: n.Name}, false
	case ast.Apply:
		left, leftChanged := rewrite(n.Left)
		right, rightChanged := rewrite(n.Right)
		current := ast.Apply{Left: left, Right: right}

		if matched, replacement := rewriteExpLog(current); matched {
			return replacement, true
		}
		if matched, replacement := rewriteExpZero(current); matched {
			return replacement, true
		}

		return current, leftChanged || rightChanged
	default:
		return expr, false
	}
}

// rewriteExpLog collapses exp(log(x)) back to x.
//
// In raw EML under the current implementation:
//   exp(x) = eml(x, 1)
//   log(x) = eml(1, eml(eml(1, x), 1))
func rewriteExpLog(expr ast.Apply) (bool, ast.Expr) {
	if !isOne(expr.Right) {
		return false, nil
	}
	left, ok := expr.Left.(ast.Apply)
	if !ok || !isOne(left.Left) {
		return false, nil
	}
	expOfOneX, ok := left.Right.(ast.Apply)
	if !ok || !isOne(expOfOneX.Right) {
		return false, nil
	}
	oneX, ok := expOfOneX.Left.(ast.Apply)
	if !ok || !isOne(oneX.Left) {
		return false, nil
	}
	return true, clone(oneX.Right)
}

// rewriteExpZero collapses exp(0) to 1, where 0 is recognized as the raw EML
// encoding of log(1).
func rewriteExpZero(expr ast.Apply) (bool, ast.Expr) {
	if !isOne(expr.Right) {
		return false, nil
	}
	if !isZero(expr.Left) {
		return false, nil
	}
	return true, ast.One{}
}

func isZero(expr ast.Expr) bool {
	zero := rawZero()
	return equal(expr, zero)
}

func rawZero() ast.Expr {
	return ast.Apply{
		Left: ast.One{},
		Right: ast.Apply{
			Left: ast.Apply{
				Left:  ast.One{},
				Right: ast.One{},
			},
			Right: ast.One{},
		},
	}
}

func isOne(expr ast.Expr) bool {
	_, ok := expr.(ast.One)
	return ok
}

func equal(a, b ast.Expr) bool {
	switch av := a.(type) {
	case ast.One:
		_, ok := b.(ast.One)
		return ok
	case ast.Variable:
		bv, ok := b.(ast.Variable)
		return ok && av.Name == bv.Name
	case ast.Apply:
		bv, ok := b.(ast.Apply)
		return ok && equal(av.Left, bv.Left) && equal(av.Right, bv.Right)
	default:
		return false
	}
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
