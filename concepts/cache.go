package concepts

import "eml-parser/ast"

type expansionCache struct {
	symbolic map[string]ast.Expr
}

func newExpansionCache() expansionCache {
	return expansionCache{
		symbolic: make(map[string]ast.Expr),
	}
}

func (c *expansionCache) clear() {
	c.symbolic = make(map[string]ast.Expr)
}

func (c *expansionCache) getSymbolic(name string) (ast.Expr, bool) {
	expr, ok := c.symbolic[name]
	if !ok {
		return nil, false
	}
	return cloneAST(expr), true
}

func (c *expansionCache) putSymbolic(name string, expr ast.Expr) {
	c.symbolic[name] = cloneAST(expr)
}
