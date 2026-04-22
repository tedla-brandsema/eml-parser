package search

import (
	"sort"

	"eml-parser/eval"
)

// SearchOptions configures the first minimal search workflow.
type SearchOptions struct {
	Bounds Bounds
	TopN   int
}

// SearchResult is one scored candidate from the initial search skeleton.
type SearchResult struct {
	Candidate Candidate
	Score     float64
}

// EnumerativeRealSearch performs a small bounded enumerative search over raw EML
// candidates and scores them against a named real-valued fixture.
func EnumerativeRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) ([]SearchResult, error) {
	atoms := AtomicSeeds(fixture.TargetKey)
	exprs := EnumerateBounded(atoms, options.Bounds)
	candidates := UniqueCandidates(exprs)

	results := make([]SearchResult, 0, len(candidates))
	for _, candidate := range candidates {
		score, err := RealMSE(candidate, backend, fixture.Samples)
		if err != nil {
			continue
		}
		results = append(results, SearchResult{
			Candidate: candidate,
			Score:     score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Candidate.Key < results[j].Candidate.Key
		}
		return results[i].Score < results[j].Score
	})

	if options.TopN > 0 && len(results) > options.TopN {
		results = results[:options.TopN]
	}
	return results, nil
}
