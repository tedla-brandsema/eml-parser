package search

import (
	"fmt"
	"math"
	"sort"

	"eml-parser/ast"
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

// LayerDiagnostics records per-depth statistics from a layered search run.
type LayerDiagnostics struct {
	Depth          int
	CandidateCount int
	BestScore      float64
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
	Layers                []LayerDiagnostics // populated by LayeredRealSearch; nil otherwise
}

// SearchReport is the inspectable output of a search run.
type SearchReport struct {
	Results     []SearchResult
	Diagnostics SearchDiagnostics
}

func (d SearchDiagnostics) String() string {
	s := fmt.Sprintf(
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
	if len(d.Layers) > 0 {
		s += fmt.Sprintf("\nlayers: %d (depths %d–%d)",
			len(d.Layers),
			d.Layers[0].Depth,
			d.Layers[len(d.Layers)-1].Depth,
		)
	}
	return s
}

// EnumerativeRealSearch performs a small bounded enumerative search over raw EML
// candidates and scores them against a named real-valued fixture.
func EnumerativeRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	target := NewSearchTarget([]string{fixture.TargetKey}, fixture.Samples)
	return EnumerativeRealSearchWithPolicies(target, backend, options, RealMSEScorer{}, RankedFullMatchPolicy{})
}

// EnumerativeRealSearchWithPolicies performs bounded enumerative search over a
// real-valued target using pluggable scoring and retention policies.
func EnumerativeRealSearchWithPolicies(target SearchTarget[float64], backend eval.Backend[complex128], options SearchOptions, scorer Scorer[float64], retention RetentionPolicy) (SearchReport, error) {
	atoms := AtomicSeeds(target.VariableNames()...)
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
		scored, err := scorer.ScoreCandidate(candidate, backend, target)
		if err != nil {
			diagnostics.EvaluationRejects++
			continue
		}
		if !scored.Finite {
			diagnostics.NonFiniteCount++
			continue
		}
		outcome := retention.Decide(RetentionContext{Current: scored})
		if outcome.Decision != RetentionContinue {
			continue
		}
		results = append(results, SearchResult{
			Candidate: candidate,
			Score:     scored.Primary,
		})
		totalScore += scored.Primary
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

// LayeredRealSearch performs a depth-by-depth bounded search over raw EML
// candidates, stopping early when a near-zero score (< 1e-12) is found.
// Per-depth statistics are recorded in SearchDiagnostics.Layers.
func LayeredRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	target := NewSearchTarget([]string{fixture.TargetKey}, fixture.Samples)
	return LayeredRealSearchWithPolicies(target, backend, options, RealMSEScorer{}, RankedFullMatchPolicy{})
}

// LayeredRealSearchWithPolicies performs depth-by-depth bounded search over a
// real-valued target using pluggable scoring and retention policies.
func LayeredRealSearchWithPolicies(target SearchTarget[float64], backend eval.Backend[complex128], options SearchOptions, scorer Scorer[float64], retention RetentionPolicy) (SearchReport, error) {
	atoms := AtomicSeeds(target.VariableNames()...)

	maxDepth := options.Bounds.MaxDepth
	if maxDepth == 0 {
		maxDepth = 8
	}

	// allRaw accumulates every raw expression discovered across all layers;
	// used as input to EnumerateNextLayer for building the next depth.
	allRaw := make([]ast.Expr, len(atoms))
	copy(allRaw, atoms)
	currentLayer := allRaw

	// scoredKeys prevents re-scoring a normalized canonical key seen in a prior layer.
	scoredKeys := make(map[string]bool)

	var allResults []SearchResult
	globalBest := math.Inf(1)
	diagnostics := SearchDiagnostics{}

	for depth := 1; depth <= maxDepth; depth++ {
		// Build unique layer candidates, skipping already-scored keys.
		var layerCandidates []Candidate
		for _, expr := range currentLayer {
			c := NewCandidate(expr)
			if !scoredKeys[c.Key] {
				scoredKeys[c.Key] = true
				layerCandidates = append(layerCandidates, c)
			}
		}

		diagnostics.GeneratedCount += len(currentLayer)
		diagnostics.UniqueCount += len(layerCandidates)
		diagnostics.DuplicateCount += len(currentLayer) - len(layerCandidates)
		for _, c := range layerCandidates {
			if !Equal(c.Original, c.Normalized) {
				diagnostics.NormalizationHits++
			}
		}

		// Score layer candidates.
		layerBest := math.Inf(1)
		for _, candidate := range layerCandidates {
			scored, err := scorer.ScoreCandidate(candidate, backend, target)
			if err != nil {
				diagnostics.EvaluationRejects++
				continue
			}
			if !scored.Finite {
				diagnostics.NonFiniteCount++
				continue
			}
			outcome := retention.Decide(RetentionContext{Current: scored})
			if outcome.Decision != RetentionContinue {
				continue
			}
			allResults = append(allResults, SearchResult{Candidate: candidate, Score: scored.Primary})
			diagnostics.ScoredCount++
			if scored.Primary < layerBest {
				layerBest = scored.Primary
			}
		}

		layerBestRecorded := layerBest
		if math.IsInf(layerBestRecorded, 1) {
			layerBestRecorded = 0
		}
		diagnostics.Layers = append(diagnostics.Layers, LayerDiagnostics{
			Depth:          depth,
			CandidateCount: len(layerCandidates),
			BestScore:      layerBestRecorded,
		})

		if !math.IsInf(layerBest, 1) && layerBest < globalBest {
			globalBest = layerBest
		}
		if globalBest < 1e-12 {
			break
		}

		nextLayer := EnumerateNextLayer(currentLayer, allRaw, options.Bounds)
		allRaw = append(allRaw, nextLayer...)
		currentLayer = nextLayer
		if len(currentLayer) == 0 {
			break
		}
	}

	// Compute global score statistics.
	if len(allResults) > 0 {
		var totalScore float64
		diagnostics.BestScore = allResults[0].Score
		diagnostics.WorstScore = allResults[0].Score
		for _, r := range allResults {
			totalScore += r.Score
			if r.Score < diagnostics.BestScore {
				diagnostics.BestScore = r.Score
			}
			if r.Score > diagnostics.WorstScore {
				diagnostics.WorstScore = r.Score
			}
		}
		diagnostics.MeanScore = totalScore / float64(len(allResults))
	}

	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].Score == allResults[j].Score {
			return allResults[i].Candidate.Key < allResults[j].Candidate.Key
		}
		return allResults[i].Score < allResults[j].Score
	})

	if options.TopN > 0 && len(allResults) > options.TopN {
		allResults = allResults[:options.TopN]
	}
	diagnostics.ReturnedCount = len(allResults)
	for _, r := range allResults {
		diagnostics.TopCandidateSummaries = append(diagnostics.TopCandidateSummaries,
			fmt.Sprintf("score=%g expr=%s", r.Score, r.Candidate.Normalized))
	}

	return SearchReport{Results: allResults, Diagnostics: diagnostics}, nil
}

func isFiniteScore(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
