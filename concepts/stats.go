package concepts

import (
	"fmt"

	"eml-parser/ast"
)

// ExpandedStats summarizes the size and dependency profile of a concept after
// symbolic expansion to raw EML.
type ExpandedStats struct {
	Concept               string
	NodeCount             int
	TreeDepth             int
	LeafCount             int
	DirectDependencyCount int
	TransitiveDepCount    int
}

// Stats expands a concept symbolically and computes structural metrics on the
// resulting raw EML tree.
func (r *Registry) Stats(name string) (ExpandedStats, error) {
	expr, err := r.ExpandSymbolic(name)
	if err != nil {
		return ExpandedStats{}, err
	}
	direct, err := r.DirectDependencies(name)
	if err != nil {
		return ExpandedStats{}, err
	}
	transitive, err := r.TransitiveDependencies(name)
	if err != nil {
		return ExpandedStats{}, err
	}

	nodes, depth, leaves := astStats(expr)
	return ExpandedStats{
		Concept:               name,
		NodeCount:             nodes,
		TreeDepth:             depth,
		LeafCount:             leaves,
		DirectDependencyCount: len(direct),
		TransitiveDepCount:    len(transitive),
	}, nil
}

func astStats(expr ast.Expr) (nodes, depth, leaves int) {
	switch n := expr.(type) {
	case ast.One, ast.Variable:
		return 1, 1, 1
	case ast.Apply:
		leftNodes, leftDepth, leftLeaves := astStats(n.Left)
		rightNodes, rightDepth, rightLeaves := astStats(n.Right)
		return 1 + leftNodes + rightNodes, 1 + max(leftDepth, rightDepth), leftLeaves + rightLeaves
	default:
		return 0, 0, 0
	}
}

func (s ExpandedStats) String() string {
	return fmt.Sprintf(
		"concept: %s\nnodes: %d\ndepth: %d\nleaves: %d\ndirect_dependencies: %d\ntransitive_dependencies: %d",
		s.Concept,
		s.NodeCount,
		s.TreeDepth,
		s.LeafCount,
		s.DirectDependencyCount,
		s.TransitiveDepCount,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
