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

func EnumerativeRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	return base.EnumerativeRealSearch(fixture, backend, options)
}

func LayeredRealSearch(fixture BenchmarkCase[float64], backend eval.Backend[complex128], options SearchOptions) (SearchReport, error) {
	return base.LayeredRealSearch(fixture, backend, options)
}
