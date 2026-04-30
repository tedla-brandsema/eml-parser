package search

import (
	"fmt"
	"math"
	"sort"

	"eml-parser/eval"
)

// SearchTarget is the evidence surface a scorer evaluates against.
// Different search routes may use different target implementations while
// sharing the same search algorithms.
type SearchTarget[T any] interface {
	Samples() []Sample[T]
	VariableNames() []string
}

// StaticTarget is the default in-memory search target for fixture-backed
// search runs.
type StaticTarget[T any] struct {
	Variables []string
	Values    []Sample[T]
}

func NewSearchTarget[T any](variables []string, samples []Sample[T]) StaticTarget[T] {
	return StaticTarget[T]{
		Variables: append([]string(nil), variables...),
		Values:    append([]Sample[T](nil), samples...),
	}
}

func (t StaticTarget[T]) Samples() []Sample[T] {
	return append([]Sample[T](nil), t.Values...)
}

func (t StaticTarget[T]) VariableNames() []string {
	return append([]string(nil), t.Variables...)
}

// ScoreResult is the structured output from scoring one candidate against one
// target. Extra fields can be added later for coverage-aware or ML-guided
// search without changing algorithm identities.
type ScoreResult struct {
	Primary          float64
	Finite           bool
	WindowStart      int
	WindowEnd        int
	CoveredCount     int
	CoverageRatio    float64
	LocalError       float64
	RawCoverageScore float64
}

type Scorer[T any] interface {
	ScoreCandidate(candidate Candidate, backend eval.Backend[complex128], target SearchTarget[T]) (ScoreResult, error)
}

type RealMSEScorer struct{}

func (RealMSEScorer) ScoreCandidate(candidate Candidate, backend eval.Backend[complex128], target SearchTarget[float64]) (ScoreResult, error) {
	score, err := RealMSE(candidate, backend, target.Samples())
	if err != nil {
		return ScoreResult{}, err
	}
	return ScoreResult{
		Primary:          score,
		Finite:           isFiniteScore(score),
		WindowStart:      0,
		WindowEnd:        len(target.Samples()),
		CoveredCount:     len(target.Samples()),
		CoverageRatio:    fullCoverageRatio(len(target.Samples())),
		LocalError:       score,
		RawCoverageScore: fullCoverageRatio(len(target.Samples())),
	}, nil
}

type ComplexMSEScorer struct{}

func (ComplexMSEScorer) ScoreCandidate(candidate Candidate, backend eval.Backend[complex128], target SearchTarget[complex128]) (ScoreResult, error) {
	score, err := ComplexMSE(candidate, backend, target.Samples())
	if err != nil {
		return ScoreResult{}, err
	}
	return ScoreResult{
		Primary:          score,
		Finite:           isFiniteScore(score),
		WindowStart:      0,
		WindowEnd:        len(target.Samples()),
		CoveredCount:     len(target.Samples()),
		CoverageRatio:    fullCoverageRatio(len(target.Samples())),
		LocalError:       score,
		RawCoverageScore: fullCoverageRatio(len(target.Samples())),
	}, nil
}

type PartialCoverageOptions struct {
	MinWindowSize  int
	MaxWindowSize  int
	CoverageWeight float64
}

type TraceWindowDiagnostics struct {
	WindowsEvaluated int
	ShapeRejects     int
}

type RealPartialCoverageScorer struct {
	Options PartialCoverageOptions
}

func (s RealPartialCoverageScorer) ScoreCandidate(candidate Candidate, backend eval.Backend[complex128], target SearchTarget[float64]) (ScoreResult, error) {
	opts := s.Options
	opts = normalizePartialCoverageOptions(opts)
	variables := target.VariableNames()
	if len(variables) != 1 {
		return ScoreResult{}, fmt.Errorf("partial coverage scorer requires exactly one variable, got %d", len(variables))
	}
	samples := sortedRealSamples(target.Samples(), variables[0])
	n := len(samples)
	if n < opts.MinWindowSize {
		return ScoreResult{}, fmt.Errorf("partial coverage scorer requires at least %d samples, got %d", opts.MinWindowSize, n)
	}
	if opts.MaxWindowSize <= 0 || opts.MaxWindowSize > n {
		opts.MaxWindowSize = n
	}

	preds := make([]float64, 0, n)
	targets := make([]float64, 0, n)
	for _, sample := range samples {
		vars := map[string]complex128{
			variables[0]: complex(sample.Vars[variables[0]], 0),
		}
		got, err := eval.EvaluateMap(candidate.Normalized, backend, vars)
		if err != nil {
			return ScoreResult{}, err
		}
		preds = append(preds, real(got))
		targets = append(targets, sample.Target)
	}

	best := ScoreResult{Primary: 0, Finite: false}
	bestInitialized := false
	for start := 0; start < n; start++ {
		maxEnd := min(n, start+opts.MaxWindowSize)
		for end := start + opts.MinWindowSize; end <= maxEnd; end++ {
			localError := windowMSE(preds[start:end], targets[start:end])
			current := coverageScoreResult(start, end, n, localError, opts.CoverageWeight)
			if !current.Finite {
				continue
			}
			if !bestInitialized || scoreResultLess(current, best) {
				best = current
				bestInitialized = true
			}
		}
	}
	if !bestInitialized {
		return ScoreResult{Finite: false}, nil
	}
	return best, nil
}

// ScoreAlignedTraceWindows scores a candidate trace against contiguous windows
// of a larger target trace using the same fit-plus-coverage objective as the
// real partial-coverage scorer. Windows must align exactly on x coordinates.
func ScoreAlignedTraceWindows(target, candidate [][2]float64, options PartialCoverageOptions) (ScoreResult, TraceWindowDiagnostics, bool) {
	opts := normalizePartialCoverageOptions(options)
	diag := TraceWindowDiagnostics{}
	if len(candidate) == 0 || len(target) < len(candidate) {
		return ScoreResult{}, diag, false
	}
	if len(candidate) < opts.MinWindowSize {
		return ScoreResult{}, diag, false
	}
	if opts.MaxWindowSize > 0 && len(candidate) > opts.MaxWindowSize {
		return ScoreResult{}, diag, false
	}

	best := ScoreResult{}
	bestInitialized := false
	windowSize := len(candidate)
	for start := 0; start+windowSize <= len(target); start++ {
		end := start + windowSize
		diag.WindowsEvaluated++
		window := target[start:end]
		localError, ok := alignedTraceMSE(window, candidate)
		if !ok {
			diag.ShapeRejects++
			continue
		}
		current := coverageScoreResult(start, end, len(target), localError, opts.CoverageWeight)
		if !current.Finite {
			continue
		}
		if !bestInitialized || scoreResultLess(current, best) {
			best = current
			bestInitialized = true
		}
	}
	if !bestInitialized {
		return ScoreResult{}, diag, false
	}
	return best, diag, true
}

type RetentionDecision string

const (
	RetentionContinue      RetentionDecision = "continue"
	RetentionRetainPartial RetentionDecision = "retain_partial"
	RetentionPrune         RetentionDecision = "prune"
)

type RetentionOutcome struct {
	Decision RetentionDecision
	Reason   string
}

type RetentionContext struct {
	Parent  *ScoreResult
	Current ScoreResult
}

type RetentionPolicy interface {
	Decide(RetentionContext) RetentionOutcome
}

// RankedFullMatchPolicy preserves the current enumerative behavior: every
// finite scored candidate remains eligible for ranking and none are retained
// separately as partials.
type RankedFullMatchPolicy struct{}

func (RankedFullMatchPolicy) Decide(ctx RetentionContext) RetentionOutcome {
	if !ctx.Current.Finite {
		return RetentionOutcome{Decision: RetentionPrune, Reason: "non_finite"}
	}
	return RetentionOutcome{Decision: RetentionContinue, Reason: "ranked_full_match"}
}

// ThresholdRetentionPolicy preserves the current maze behavior while keeping it
// outside the maze algorithm itself.
type ThresholdRetentionPolicy struct {
	AcceptThreshold float64
	RetainThreshold float64
	MinImprovement  float64
}

func (p ThresholdRetentionPolicy) Decide(ctx RetentionContext) RetentionOutcome {
	if !ctx.Current.Finite {
		return RetentionOutcome{Decision: RetentionPrune, Reason: "non_finite"}
	}
	improvement := 0.0
	if ctx.Parent != nil {
		improvement = ctx.Parent.Primary - ctx.Current.Primary
	}
	switch {
	case ctx.Current.Primary <= p.AcceptThreshold && improvement >= p.MinImprovement:
		return RetentionOutcome{Decision: RetentionContinue, Reason: "accepted"}
	case ctx.Current.Primary <= p.RetainThreshold:
		if improvement < p.MinImprovement {
			return RetentionOutcome{Decision: RetentionRetainPartial, Reason: "stalled"}
		}
		return RetentionOutcome{Decision: RetentionRetainPartial, Reason: "retained_after_validation"}
	default:
		return RetentionOutcome{Decision: RetentionPrune, Reason: "score_above_retain_threshold"}
	}
}

type CoverageRetentionPolicy struct {
	AcceptThreshold float64
	RetainThreshold float64
	MinImprovement  float64
	MinCoveredCount int
}

func (p CoverageRetentionPolicy) Decide(ctx RetentionContext) RetentionOutcome {
	if !ctx.Current.Finite {
		return RetentionOutcome{Decision: RetentionPrune, Reason: "non_finite"}
	}
	if p.MinCoveredCount > 0 && ctx.Current.CoveredCount < p.MinCoveredCount {
		return RetentionOutcome{Decision: RetentionPrune, Reason: "coverage_too_small"}
	}
	improvement := 0.0
	if ctx.Parent != nil {
		improvement = ctx.Parent.Primary - ctx.Current.Primary
	}
	switch {
	case ctx.Current.Primary <= p.AcceptThreshold && improvement >= p.MinImprovement:
		return RetentionOutcome{Decision: RetentionContinue, Reason: "accepted_with_coverage"}
	case ctx.Current.Primary <= p.RetainThreshold:
		if improvement < p.MinImprovement {
			return RetentionOutcome{Decision: RetentionRetainPartial, Reason: "stalled_with_coverage"}
		}
		return RetentionOutcome{Decision: RetentionRetainPartial, Reason: "retained_partial_coverage"}
	default:
		return RetentionOutcome{Decision: RetentionPrune, Reason: "score_above_retain_threshold"}
	}
}

func sortedRealSamples(samples []Sample[float64], variable string) []Sample[float64] {
	out := append([]Sample[float64](nil), samples...)
	sort.Slice(out, func(i, j int) bool {
		ix := out[i].Vars[variable]
		jx := out[j].Vars[variable]
		if ix == jx {
			return out[i].Target < out[j].Target
		}
		return ix < jx
	})
	return out
}

func windowMSE(preds, targets []float64) float64 {
	var total float64
	for i := range preds {
		diff := preds[i] - targets[i]
		total += diff * diff
	}
	return total / float64(len(preds))
}

func alignedTraceMSE(target, candidate [][2]float64) (float64, bool) {
	if len(target) == 0 || len(target) != len(candidate) {
		return 0, false
	}
	var total float64
	for i := range target {
		if math.Abs(target[i][0]-candidate[i][0]) > 1e-9 {
			return 0, false
		}
		diff := target[i][1] - candidate[i][1]
		total += diff * diff
	}
	return total / float64(len(target)), true
}

func scoreResultLess(a, b ScoreResult) bool {
	if a.Primary == b.Primary {
		if a.CoveredCount == b.CoveredCount {
			if a.LocalError == b.LocalError {
				return a.WindowStart < b.WindowStart
			}
			return a.LocalError < b.LocalError
		}
		return a.CoveredCount > b.CoveredCount
	}
	return a.Primary < b.Primary
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func fullCoverageRatio(count int) float64 {
	if count <= 0 {
		return 0
	}
	return 1
}

func normalizePartialCoverageOptions(opts PartialCoverageOptions) PartialCoverageOptions {
	if opts.MinWindowSize <= 0 {
		opts.MinWindowSize = 3
	}
	if opts.CoverageWeight == 0 {
		opts.CoverageWeight = 0.25
	}
	return opts
}

func coverageScoreResult(start, end, total int, localError float64, coverageWeight float64) ScoreResult {
	covered := end - start
	coverageRatio := 0.0
	if total > 0 {
		coverageRatio = float64(covered) / float64(total)
	}
	primary := localError + coverageWeight*(1.0-coverageRatio)
	return ScoreResult{
		Primary:          primary,
		Finite:           isFiniteScore(primary),
		WindowStart:      start,
		WindowEnd:        end,
		CoveredCount:     covered,
		CoverageRatio:    coverageRatio,
		LocalError:       localError,
		RawCoverageScore: coverageRatio,
	}
}
