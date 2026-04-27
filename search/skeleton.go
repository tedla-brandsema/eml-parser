package search

import (
	"fmt"
	"math"
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

// SearchDiagnostics summarizes how a search run generated, reduced, and scored
// candidates.
type SearchDiagnostics struct {
	GeneratedCount        int
	UniqueCount           int
	DuplicateCount        int
	NormalizationHits     int
	EvaluationRejects     int
	NonFiniteCount        int
	ScoredCount           int
	ReturnedCount         int
	BestScore             float64
	WorstScore            float64
	MeanScore             float64
	TopCandidateSummaries []string
}

// SearchReport is the inspectable output of a search run.
type SearchReport struct {
	Results     []SearchResult
	Diagnostics SearchDiagnostics
}

func (d SearchDiagnostics) String() string {
	return fmt.Sprintf(
		"generated: %d\nunique: %d\nduplicates: %d\nnormalization_hits: %d\nevaluation_rejects: %d\nnon_finite_count: %d\nscored: %d\nreturned: %d\nbest_score: %g\nworst_score: %g\nmean_score: %g",
		d.GeneratedCount,
		d.UniqueCount,
		d.DuplicateCount,
		d.NormalizationHits,
		d.EvaluationRejects,
		d.NonFiniteCount,
		d.ScoredCount,
		d.ReturnedCount,
		d.BestScore,
		d.WorstScore,
		d.MeanScore,
	)
}

// EnumerativeRealSearch performs a small bounded enumerative search over raw EML
// candidates and scores them against a named real-valued fixture.
func EnumerativeRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	atoms := AtomicSeeds(fixture.TargetKey)
	exprs := EnumerateBounded(atoms, options.Bounds)
	candidates := UniqueCandidates(exprs)

	diagnostics := SearchDiagnostics{
		GeneratedCount: len(exprs),
		UniqueCount:    len(candidates),
		DuplicateCount: len(exprs) - len(candidates),
	}
	for _, candidate := range candidates {
		if !Equal(candidate.Original, candidate.Normalized) {
			diagnostics.NormalizationHits++
		}
	}

	results := make([]SearchResult, 0, len(candidates))
	var totalScore float64
	for _, candidate := range candidates {
		score, err := RealMSE(candidate, backend, fixture.Samples)
		if err != nil {
			diagnostics.EvaluationRejects++
			continue
		}
		if !isFiniteScore(score) {
			diagnostics.NonFiniteCount++
			continue
		}
		results = append(results, SearchResult{
			Candidate: candidate,
			Score:     score,
		})
		totalScore += score
	}
	diagnostics.ScoredCount = len(results)
	if diagnostics.ScoredCount > 0 {
		diagnostics.BestScore = results[0].Score
		diagnostics.WorstScore = results[0].Score
		diagnostics.MeanScore = totalScore / float64(diagnostics.ScoredCount)
		for _, result := range results {
			if result.Score < diagnostics.BestScore {
				diagnostics.BestScore = result.Score
			}
			if result.Score > diagnostics.WorstScore {
				diagnostics.WorstScore = result.Score
			}
		}
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
	diagnostics.ReturnedCount = len(results)
	for _, result := range results {
		diagnostics.TopCandidateSummaries = append(
			diagnostics.TopCandidateSummaries,
			fmt.Sprintf("score=%g expr=%s", result.Score, result.Candidate.Normalized),
		)
	}
	return SearchReport{
		Results:     results,
		Diagnostics: diagnostics,
	}, nil
}

func isFiniteScore(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
