package search

import (
	"eml-parser/ast"
	"eml-parser/normalize"
)

// Candidate is a raw-EML search unit with normalized comparison metadata.
type Candidate struct {
	Original   ast.Expr
	Normalized ast.Expr
	Key        string
	Stats      Stats
}

// NewCandidate constructs a candidate from a raw EML AST.
func NewCandidate(expr ast.Expr) Candidate {
	original := clone(expr)
	normalized := normalize.Expr(clone(expr))
	return Candidate{
		Original:   original,
		Normalized: normalized,
		Key:        CanonicalKey(normalized),
		Stats:      TreeStats(normalized),
	}
}
