package search

import "eml-parser/eval"

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
	Primary float64
	Finite  bool
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
		Primary: score,
		Finite:  isFiniteScore(score),
	}, nil
}

type ComplexMSEScorer struct{}

func (ComplexMSEScorer) ScoreCandidate(candidate Candidate, backend eval.Backend[complex128], target SearchTarget[complex128]) (ScoreResult, error) {
	score, err := ComplexMSE(candidate, backend, target.Samples())
	if err != nil {
		return ScoreResult{}, err
	}
	return ScoreResult{
		Primary: score,
		Finite:  isFiniteScore(score),
	}, nil
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
