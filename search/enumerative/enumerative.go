package enumerative

import (
	"eml-parser/eval"
	base "eml-parser/search"
)

type SearchOptions = base.SearchOptions
type SearchResult = base.SearchResult
type SearchReport = base.SearchReport
type SearchDiagnostics = base.SearchDiagnostics
type LayerDiagnostics = base.LayerDiagnostics
type BenchmarkCase[T any] = base.BenchmarkCase[T]
type SearchTarget[T any] = base.SearchTarget[T]
type Scorer[T any] = base.Scorer[T]
type RetentionPolicy = base.RetentionPolicy

func EnumerativeRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	return base.EnumerativeRealSearch(fixture, backend, options)
}

func EnumerativeRealSearchWithPolicies(target SearchTarget[float64], backend eval.Backend[complex128], options SearchOptions, scorer Scorer[float64], retention RetentionPolicy) (SearchReport, error) {
	return base.EnumerativeRealSearchWithPolicies(target, backend, options, scorer, retention)
}

func LayeredRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	return base.LayeredRealSearch(fixture, backend, options)
}

func LayeredRealSearchWithPolicies(target SearchTarget[float64], backend eval.Backend[complex128], options SearchOptions, scorer Scorer[float64], retention RetentionPolicy) (SearchReport, error) {
	return base.LayeredRealSearchWithPolicies(target, backend, options, scorer, retention)
}
