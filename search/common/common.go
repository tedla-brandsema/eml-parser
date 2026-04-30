package common

import (
	"eml-parser/ast"
	"eml-parser/eval"
	base "eml-parser/search"
)

type Candidate = base.Candidate
type Stats = base.Stats
type Bounds = base.Bounds
type Sample[T any] = base.Sample[T]
type BenchmarkCase[T any] = base.BenchmarkCase[T]
type SearchTarget[T any] = base.SearchTarget[T]
type StaticTarget[T any] = base.StaticTarget[T]
type ScoreResult = base.ScoreResult
type Scorer[T any] = base.Scorer[T]
type RetentionDecision = base.RetentionDecision
type RetentionOutcome = base.RetentionOutcome
type RetentionContext = base.RetentionContext
type RetentionPolicy = base.RetentionPolicy
type RealMSEScorer = base.RealMSEScorer
type ComplexMSEScorer = base.ComplexMSEScorer
type RankedFullMatchPolicy = base.RankedFullMatchPolicy
type ThresholdRetentionPolicy = base.ThresholdRetentionPolicy

const (
	RetentionContinue      = base.RetentionContinue
	RetentionRetainPartial = base.RetentionRetainPartial
	RetentionPrune         = base.RetentionPrune
)

func NewCandidate(expr ast.Expr) Candidate { return base.NewCandidate(expr) }
func TreeStats(expr ast.Expr) Stats        { return base.TreeStats(expr) }
func CanonicalKey(expr ast.Expr) string    { return base.CanonicalKey(expr) }
func Equal(a, b ast.Expr) bool             { return base.Equal(a, b) }
func Subtrees(expr ast.Expr) []ast.Expr    { return base.Subtrees(expr) }
func NewSearchTarget[T any](variables []string, samples []Sample[T]) StaticTarget[T] {
	return base.NewSearchTarget(variables, samples)
}

func AtomicSeeds(variableNames ...string) []ast.Expr {
	return base.AtomicSeeds(variableNames...)
}

func WithinBounds(expr ast.Expr, bounds Bounds) bool {
	return base.WithinBounds(expr, bounds)
}

func UniqueCandidates(exprs []ast.Expr) []Candidate {
	return base.UniqueCandidates(exprs)
}

func DeduplicateCandidates(candidates []Candidate) []Candidate {
	return base.DeduplicateCandidates(candidates)
}

func ReplaceSubtree(expr ast.Expr, preorderIndex int, replacement ast.Expr) (ast.Expr, error) {
	return base.ReplaceSubtree(expr, preorderIndex, replacement)
}

func MutateByReplacement(expr ast.Expr, replacements []ast.Expr, bounds Bounds) []Candidate {
	return base.MutateByReplacement(expr, replacements, bounds)
}

func RealMSE(candidate Candidate, backend eval.Backend[complex128], samples []Sample[float64]) (float64, error) {
	return base.RealMSE(candidate, backend, samples)
}

func ComplexMSE(candidate Candidate, backend eval.Backend[complex128], samples []Sample[complex128]) (float64, error) {
	return base.ComplexMSE(candidate, backend, samples)
}

func ComplexRMSE(candidate Candidate, backend eval.Backend[complex128], samples []Sample[complex128]) (float64, error) {
	return base.ComplexRMSE(candidate, backend, samples)
}

func RealBenchmarkFixtures() ([]BenchmarkCase[float64], error) {
	return base.RealBenchmarkFixtures()
}

func ComplexBenchmarkFixtures() ([]BenchmarkCase[complex128], error) {
	return base.ComplexBenchmarkFixtures()
}
