package search

import (
	"fmt"

	"eml-parser/ast"
)

// Bounds constrain generated or rewritten raw EML trees.
type Bounds struct {
	MaxDepth int
	MaxNodes int
}

// WithinBounds reports whether an expression satisfies the requested limits.
// Zero-valued limits are treated as "unbounded".
func WithinBounds(expr ast.Expr, bounds Bounds) bool {
	stats := TreeStats(expr)
	if bounds.MaxDepth > 0 && stats.TreeDepth > bounds.MaxDepth {
		return false
	}
	if bounds.MaxNodes > 0 && stats.NodeCount > bounds.MaxNodes {
		return false
	}
	return true
}

// ReplaceSubtree replaces the subtree at the given pre-order index.
func ReplaceSubtree(expr ast.Expr, preorderIndex int, replacement ast.Expr) (ast.Expr, error) {
	if preorderIndex < 0 {
		return nil, fmt.Errorf("preorder index must be non-negative")
	}
	nextIndex := 0
	out, replaced := replaceSubtree(expr, preorderIndex, replacement, &nextIndex)
	if !replaced {
		return nil, fmt.Errorf("preorder index %d out of range", preorderIndex)
	}
	return out, nil
}

func replaceSubtree(expr ast.Expr, target int, replacement ast.Expr, nextIndex *int) (ast.Expr, bool) {
	current := *nextIndex
	*nextIndex = *nextIndex + 1
	if current == target {
		return clone(replacement), true
	}

	switch n := expr.(type) {
	case ast.One:
		return ast.One{}, false
	case ast.Variable:
		return ast.Variable{Name: n.Name}, false
	case ast.Apply:
		left, replaced := replaceSubtree(n.Left, target, replacement, nextIndex)
		if replaced {
			return ast.Apply{Left: left, Right: clone(n.Right)}, true
		}
		right, replaced := replaceSubtree(n.Right, target, replacement, nextIndex)
		if replaced {
			return ast.Apply{Left: clone(n.Left), Right: right}, true
		}
		return ast.Apply{Left: left, Right: right}, false
	default:
		return expr, false
	}
}

// MutateByReplacement replaces each subtree with each replacement and returns
// unique normalized candidates that satisfy the given bounds.
func MutateByReplacement(expr ast.Expr, replacements []ast.Expr, bounds Bounds) []Candidate {
	subtrees := Subtrees(expr)
	var generated []ast.Expr
	for idx := range subtrees {
		for _, replacement := range replacements {
			mutated, err := ReplaceSubtree(expr, idx, replacement)
			if err != nil {
				continue
			}
			if !WithinBounds(mutated, bounds) {
				continue
			}
			generated = append(generated, mutated)
		}
	}
	return UniqueCandidates(generated)
}
