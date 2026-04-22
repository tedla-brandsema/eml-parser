package search

import "eml-parser/ast"

// UniqueCandidates normalizes and deduplicates raw expressions by canonical key.
func UniqueCandidates(exprs []ast.Expr) []Candidate {
	seen := make(map[string]Candidate)
	for _, expr := range exprs {
		candidate := NewCandidate(expr)
		if _, ok := seen[candidate.Key]; ok {
			continue
		}
		seen[candidate.Key] = candidate
	}

	out := make([]Candidate, 0, len(seen))
	for _, candidate := range seen {
		out = append(out, candidate)
	}
	return out
}

// DeduplicateCandidates removes duplicate candidates by canonical key.
func DeduplicateCandidates(candidates []Candidate) []Candidate {
	seen := make(map[string]Candidate)
	for _, candidate := range candidates {
		if _, ok := seen[candidate.Key]; ok {
			continue
		}
		seen[candidate.Key] = candidate
	}
	out := make([]Candidate, 0, len(seen))
	for _, candidate := range seen {
		out = append(out, candidate)
	}
	return out
}
