package eval

import (
	"fmt"

	"eml-parser/ast"
)

// Backend evaluates the paper-grounded EML operator over a concrete numeric
// representation.
type Backend[T any] interface {
	One() T
	EML(left, right T) (T, error)
}

// Bindings resolves variable names to backend values during evaluation.
type Bindings[T any] interface {
	Lookup(name string) (T, bool)
}

// MapBindings adapts a Go map into the Bindings interface.
type MapBindings[T any] map[string]T

// Lookup resolves a variable from the map-backed environment.
func (m MapBindings[T]) Lookup(name string) (T, bool) {
	value, ok := m[name]
	return value, ok
}

// Evaluate walks an EML expression using the supplied backend and bindings.
func Evaluate[T any](expr ast.Expr, backend Backend[T], bindings Bindings[T]) (T, error) {
	switch n := expr.(type) {
	case ast.One:
		return backend.One(), nil
	case ast.Variable:
		if bindings == nil {
			var zero T
			return zero, fmt.Errorf("unbound variable %q", n.Name)
		}
		value, ok := bindings.Lookup(n.Name)
		if !ok {
			var zero T
			return zero, fmt.Errorf("unbound variable %q", n.Name)
		}
		return value, nil
	case ast.Apply:
		left, err := Evaluate(n.Left, backend, bindings)
		if err != nil {
			var zero T
			return zero, err
		}
		right, err := Evaluate(n.Right, backend, bindings)
		if err != nil {
			var zero T
			return zero, err
		}
		return backend.EML(left, right)
	default:
		var zero T
		return zero, fmt.Errorf("unsupported expression type %T", expr)
	}
}

// EvaluateMap is a convenience wrapper for map-backed variable bindings.
func EvaluateMap[T any](expr ast.Expr, backend Backend[T], vars map[string]T) (T, error) {
	return Evaluate(expr, backend, MapBindings[T](vars))
}
