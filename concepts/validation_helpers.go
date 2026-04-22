package concepts

import "eml-parser/ast"

// StatsForExpr computes structural metrics for an already-built raw EML tree.
// This supports validation work without forcing callers through named concept expansion.
func StatsForExpr(expr ast.Expr) ExpandedStats {
	nodes, depth, leaves := astStats(expr)
	return ExpandedStats{
		NodeCount: nodes,
		TreeDepth: depth,
		LeafCount: leaves,
	}
}
