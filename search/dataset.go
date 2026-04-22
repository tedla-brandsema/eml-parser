package search

import (
	"fmt"
	"math/cmplx"

	"eml-parser/ast"
	"eml-parser/concepts"
)

// BenchmarkCase is a reusable regression fixture for a named target expression.
type BenchmarkCase[T any] struct {
	Name      string
	Expr      ast.Expr
	Samples   []Sample[T]
	TargetKey string
}

// RealRangeSamples builds real-valued samples from one variable over an evenly
// spaced range.
func RealRangeSamples(varName string, start, stop float64, count int, target func(float64) float64) []Sample[float64] {
	if count <= 0 {
		return nil
	}
	if count == 1 {
		return []Sample[float64]{{
			Vars:   map[string]float64{varName: start},
			Target: target(start),
		}}
	}

	step := (stop - start) / float64(count-1)
	samples := make([]Sample[float64], 0, count)
	for i := 0; i < count; i++ {
		x := start + float64(i)*step
		samples = append(samples, Sample[float64]{
			Vars:   map[string]float64{varName: x},
			Target: target(x),
		})
	}
	return samples
}

// ComplexGridSamples builds complex-valued samples for one variable over the
// Cartesian product of the supplied real and imaginary coordinates.
func ComplexGridSamples(varName string, reals, imags []float64, target func(complex128) complex128) []Sample[complex128] {
	var samples []Sample[complex128]
	for _, re := range reals {
		for _, im := range imags {
			z := complex(re, im)
			samples = append(samples, Sample[complex128]{
				Vars:   map[string]complex128{varName: z},
				Target: target(z),
			})
		}
	}
	return samples
}

// RealBenchmarkFixtures returns a small set of reusable real-valued regression
// targets grounded in the current concept library.
func RealBenchmarkFixtures() ([]BenchmarkCase[float64], error) {
	registry := concepts.StandardLibrary()

	expExpr, err := registry.ExpandSymbolic("exp")
	if err != nil {
		return nil, fmt.Errorf("expand exp fixture: %w", err)
	}
	sigmoidExpr, err := registry.ExpandSymbolic("sigmoid")
	if err != nil {
		return nil, fmt.Errorf("expand sigmoid fixture: %w", err)
	}

	return []BenchmarkCase[float64]{
		{
			Name:      "exp_real_small",
			Expr:      expExpr,
			Samples:   RealRangeSamples("x", -1, 1, 5, func(x float64) float64 { return real(cmplx.Exp(complex(x, 0))) }),
			TargetKey: "x",
		},
		{
			Name:      "sigmoid_real_small",
			Expr:      sigmoidExpr,
			Samples:   RealRangeSamples("x", -2, 2, 5, func(x float64) float64 { return real(1 / (1 + cmplx.Exp(complex(-x, 0)))) }),
			TargetKey: "x",
		},
	}, nil
}

// ComplexBenchmarkFixtures returns a small set of reusable complex-valued
// regression targets grounded in the current concept library.
func ComplexBenchmarkFixtures() ([]BenchmarkCase[complex128], error) {
	registry := concepts.StandardLibrary()

	sinExpr, err := registry.ExpandSymbolic("sin")
	if err != nil {
		return nil, fmt.Errorf("expand sin fixture: %w", err)
	}
	logExpr, err := registry.ExpandSymbolic("log")
	if err != nil {
		return nil, fmt.Errorf("expand log fixture: %w", err)
	}

	return []BenchmarkCase[complex128]{
		{
			Name:      "sin_complex_grid",
			Expr:      sinExpr,
			Samples:   ComplexGridSamples("x", []float64{-0.5, 0, 0.5}, []float64{-0.25, 0.25}, cmplx.Sin),
			TargetKey: "x",
		},
		{
			Name:      "log_complex_grid",
			Expr:      logExpr,
			Samples:   ComplexGridSamples("x", []float64{0.5, 1.5, 2.5}, []float64{-0.5, 0.5}, cmplx.Log),
			TargetKey: "x",
		},
	}, nil
}
