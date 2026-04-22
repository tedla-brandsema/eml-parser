package search

import (
	"math"

	"eml-parser/eval"
)

// Sample is one supervised observation for a candidate expression.
type Sample[T any] struct {
	Vars   map[string]T
	Target T
}

// ComplexMSE computes mean squared error over complex-valued samples using the
// squared magnitude of the residual.
func ComplexMSE(candidate Candidate, backend eval.Backend[complex128], samples []Sample[complex128]) (float64, error) {
	if len(samples) == 0 {
		return 0, nil
	}
	var total float64
	for _, sample := range samples {
		got, err := eval.EvaluateMap(candidate.Normalized, backend, sample.Vars)
		if err != nil {
			return 0, err
		}
		diff := got - sample.Target
		total += real(diff)*real(diff) + imag(diff)*imag(diff)
	}
	return total / float64(len(samples)), nil
}

// RealMSE computes mean squared error over real-valued samples.
func RealMSE(candidate Candidate, backend eval.Backend[complex128], samples []Sample[float64]) (float64, error) {
	if len(samples) == 0 {
		return 0, nil
	}
	var total float64
	for _, sample := range samples {
		vars := make(map[string]complex128, len(sample.Vars))
		for name, value := range sample.Vars {
			vars[name] = complex(value, 0)
		}
		got, err := eval.EvaluateMap(candidate.Normalized, backend, vars)
		if err != nil {
			return 0, err
		}
		diff := real(got) - sample.Target
		total += diff * diff
	}
	return total / float64(len(samples)), nil
}

// ComplexRMSE computes root mean squared error over complex-valued samples.
func ComplexRMSE(candidate Candidate, backend eval.Backend[complex128], samples []Sample[complex128]) (float64, error) {
	mse, err := ComplexMSE(candidate, backend, samples)
	if err != nil {
		return 0, err
	}
	return math.Sqrt(mse), nil
}
