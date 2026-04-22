package concepts

import (
	"math"
	"math/cmplx"
	"testing"

	"eml-parser/ast"
	"eml-parser/eval"
	"eml-parser/normalize"
)

func TestStandardLibraryConstantE(t *testing.T) {
	registry := StandardLibrary()

	expr, err := registry.Expand("e")
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	got, err := eval.Evaluate(expr, eval.Complex128Backend{}, nil)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if !complexClose(got, complex(math.E, 0), 1e-12) {
		t.Fatalf("expected e, got %v", got)
	}
}

func TestStandardLibraryDerivedConstants(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name string
		want complex128
	}{
		{name: "one", want: complex(1, 0)},
		{name: "zero", want: complex(0, 0)},
		{name: "minus_one", want: complex(-1, 0)},
		{name: "two", want: complex(2, 0)},
		{name: "half", want: complex(0.5, 0)},
		{name: "i", want: complex(0, 1)},
		{name: "pi", want: complex(math.Pi, 0)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := registry.Expand(tc.name)
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.Evaluate(expr, eval.Complex128Backend{}, nil)
			if err != nil {
				t.Fatalf("Evaluate returned error: %v", err)
			}
			if !complexClose(got, tc.want, 1e-8) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibraryLogMatchesReference(t *testing.T) {
	registry := StandardLibrary()

	expr, err := registry.Expand("log", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	x := complex(2.5, 0.5)
	got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": x})
	if err != nil {
		t.Fatalf("EvaluateMap returned error: %v", err)
	}
	if !complexClose(got, cmplx.Log(x), 1e-10) {
		t.Fatalf("expected %v, got %v", cmplx.Log(x), got)
	}
}

func TestStandardLibraryIdentityMatchesInput(t *testing.T) {
	registry := StandardLibrary()

	expr, err := registry.Expand("id", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	x := complex(1.5, -0.25)
	got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": x})
	if err != nil {
		t.Fatalf("EvaluateMap returned error: %v", err)
	}
	if !complexClose(got, x, 1e-10) {
		t.Fatalf("expected %v, got %v", x, got)
	}
}

func TestStandardLibraryAddMulDivPowMatchReference(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name string
		args []ast.Expr
		vars map[string]complex128
		want complex128
	}{
		{
			name: "add",
			args: []ast.Expr{ast.Variable{Name: "x"}, ast.Variable{Name: "y"}},
			vars: map[string]complex128{
				"x": complex(1.25, -0.5),
				"y": complex(-0.75, 1.25),
			},
			want: complex(1.25, -0.5) + complex(-0.75, 1.25),
		},
		{
			name: "mul",
			args: []ast.Expr{ast.Variable{Name: "x"}, ast.Variable{Name: "y"}},
			vars: map[string]complex128{
				"x": complex(2.5, 0.25),
				"y": complex(0.5, -1.5),
			},
			want: complex(2.5, 0.25) * complex(0.5, -1.5),
		},
		{
			name: "div",
			args: []ast.Expr{ast.Variable{Name: "x"}, ast.Variable{Name: "y"}},
			vars: map[string]complex128{
				"x": complex(2.5, 0.25),
				"y": complex(0.5, -1.5),
			},
			want: complex(2.5, 0.25) / complex(0.5, -1.5),
		},
		{
			name: "pow",
			args: []ast.Expr{ast.Variable{Name: "x"}, ast.Variable{Name: "y"}},
			vars: map[string]complex128{
				"x": complex(1.5, 0.25),
				"y": complex(-0.25, 0.5),
			},
			want: cmplx.Exp(complex(-0.25, 0.5) * cmplx.Log(complex(1.5, 0.25))),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := registry.Expand(tc.name, tc.args...)
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, tc.vars)
			if err != nil {
				t.Fatalf("EvaluateMap returned error: %v", err)
			}
			if !complexClose(got, tc.want, 1e-9) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibraryExpandedArithmeticWorksWithHighPrecisionBackend(t *testing.T) {
	registry := StandardLibrary()

	expr, err := registry.Expand("mul", ast.Variable{Name: "x"}, ast.Variable{Name: "y"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	got, err := eval.Evaluate(expr, eval.NewHighPrecisionComplexBackend(eval.Precision{
		WorkingBits: 256,
		LogBranch:   eval.PrincipalLogBranch,
	}), eval.MapBindings[eval.HighPrecisionComplex]{
		"x": eval.NewHighPrecisionComplex(256, 2.5, 0.25),
		"y": eval.NewHighPrecisionComplex(256, 0.5, -1.5),
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if !highPrecisionClose(got, complex(2.5, 0.25)*complex(0.5, -1.5), 1e-9) {
		t.Fatalf("unexpected high-precision result: %s", got.String())
	}
}

func TestStandardLibrarySquareAndSqrtMatchReference(t *testing.T) {
	registry := StandardLibrary()

	squareExpr, err := registry.Expand("square", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand(square) returned error: %v", err)
	}
	sqrtExpr, err := registry.Expand("sqrt", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand(sqrt) returned error: %v", err)
	}

	x := complex(1.5, 0.25)
	squareGot, err := eval.EvaluateMap(squareExpr, eval.Complex128Backend{}, map[string]complex128{"x": x})
	if err != nil {
		t.Fatalf("EvaluateMap(square) returned error: %v", err)
	}
	if !complexClose(squareGot, x*x, 1e-9) {
		t.Fatalf("expected %v, got %v", x*x, squareGot)
	}

	y := complex(2.5, 0.5)
	sqrtGot, err := eval.EvaluateMap(sqrtExpr, eval.Complex128Backend{}, map[string]complex128{"x": y})
	if err != nil {
		t.Fatalf("EvaluateMap(sqrt) returned error: %v", err)
	}
	if !complexClose(sqrtGot, cmplx.Sqrt(y), 1e-8) {
		t.Fatalf("expected %v, got %v", cmplx.Sqrt(y), sqrtGot)
	}
}

func TestStandardLibraryHyperbolicFunctionsMatchReference(t *testing.T) {
	registry := StandardLibrary()

	x := complex(0.75, -0.5)
	cases := []struct {
		name string
		want complex128
	}{
		{name: "sinh", want: cmplx.Sinh(x)},
		{name: "cosh", want: cmplx.Cosh(x)},
		{name: "tanh", want: cmplx.Tanh(x)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := registry.Expand(tc.name, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": x})
			if err != nil {
				t.Fatalf("EvaluateMap returned error: %v", err)
			}
			if !complexClose(got, tc.want, 1e-8) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibraryTrigFunctionsMatchReference(t *testing.T) {
	registry := StandardLibrary()

	x := complex(0.5, -0.25)
	cases := []struct {
		name string
		want complex128
	}{
		{name: "sin", want: cmplx.Sin(x)},
		{name: "cos", want: cmplx.Cos(x)},
		{name: "tan", want: cmplx.Tan(x)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := registry.Expand(tc.name, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": x})
			if err != nil {
				t.Fatalf("EvaluateMap returned error: %v", err)
			}
			if !complexClose(got, tc.want, 1e-7) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibraryInverseHyperbolicFunctionsMatchReference(t *testing.T) {
	registry := StandardLibrary()

	x := complex(0.4, -0.3)
	cases := []struct {
		name string
		want complex128
	}{
		{name: "asinh", want: cmplx.Asinh(x)},
		{name: "acosh", want: cmplx.Acosh(complex(1.8, 0.3))},
		{name: "atanh", want: cmplx.Atanh(x)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			arg := x
			if tc.name == "acosh" {
				arg = complex(1.8, 0.3)
			}
			expr, err := registry.Expand(tc.name, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": arg})
			if err != nil {
				t.Fatalf("EvaluateMap returned error: %v", err)
			}
			if !complexClose(got, tc.want, 1e-7) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibraryInverseTrigFunctionsMatchReference(t *testing.T) {
	registry := StandardLibrary()

	x := complex(0.35, -0.2)
	cases := []struct {
		name string
		want complex128
	}{
		{name: "asin", want: cmplx.Asin(x)},
		{name: "acos", want: cmplx.Acos(x)},
		{name: "atan", want: cmplx.Atan(x)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := registry.Expand(tc.name, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": x})
			if err != nil {
				t.Fatalf("EvaluateMap returned error: %v", err)
			}
			if !complexClose(got, tc.want, 1e-7) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibrarySigmoidMatchesReference(t *testing.T) {
	registry := StandardLibrary()

	expr, err := registry.Expand("sigmoid", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	x := complex(0.75, -0.25)
	got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": x})
	if err != nil {
		t.Fatalf("EvaluateMap returned error: %v", err)
	}
	want := 1 / (1 + cmplx.Exp(-x))
	if !complexClose(got, want, 1e-8) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestStandardLibraryInverseFunctionRoundTrips(t *testing.T) {
	cases := []struct {
		name    string
		outer   func(complex128) complex128
		inner   func(complex128) complex128
		input   complex128
		epsilon float64
	}{
		{name: "sin_asin", outer: cmplx.Sin, inner: cmplx.Asin, input: complex(0.3, -0.2), epsilon: 1e-7},
		{name: "cos_acos", outer: cmplx.Cos, inner: cmplx.Acos, input: complex(0.25, -0.1), epsilon: 1e-7},
		{name: "tan_atan", outer: cmplx.Tan, inner: cmplx.Atan, input: complex(0.2, -0.15), epsilon: 1e-7},
		{name: "sinh_asinh", outer: cmplx.Sinh, inner: cmplx.Asinh, input: complex(0.35, -0.25), epsilon: 1e-7},
		{name: "tanh_atanh", outer: cmplx.Tanh, inner: cmplx.Atanh, input: complex(0.2, -0.15), epsilon: 1e-7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.outer(tc.inner(tc.input))
			if !complexClose(want, tc.input, tc.epsilon) {
				t.Fatalf("expected round-trip to recover %v, got %v", tc.input, want)
			}
		})
	}
}

func TestStandardLibraryConceptRoundTripsMatchReference(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name    string
		outer   string
		inner   string
		input   complex128
		epsilon float64
	}{
		{name: "sin_asin", outer: "sin", inner: "asin", input: complex(0.3, -0.2), epsilon: 1e-7},
		{name: "cos_acos", outer: "cos", inner: "acos", input: complex(0.25, -0.1), epsilon: 1e-7},
		{name: "tan_atan", outer: "tan", inner: "atan", input: complex(0.2, -0.15), epsilon: 1e-7},
		{name: "sinh_asinh", outer: "sinh", inner: "asinh", input: complex(0.35, -0.25), epsilon: 1e-7},
		{name: "tanh_atanh", outer: "tanh", inner: "atanh", input: complex(0.2, -0.15), epsilon: 1e-7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			innerExpr, err := registry.Expand(tc.inner, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand inner returned error: %v", err)
			}
			innerVal, err := eval.EvaluateMap(innerExpr, eval.Complex128Backend{}, map[string]complex128{"x": tc.input})
			if err != nil {
				t.Fatalf("EvaluateMap inner returned error: %v", err)
			}

			outerExpr, err := registry.Expand(tc.outer, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand outer returned error: %v", err)
			}
			got, err := eval.EvaluateMap(outerExpr, eval.Complex128Backend{}, map[string]complex128{"x": innerVal})
			if err != nil {
				t.Fatalf("EvaluateMap outer returned error: %v", err)
			}
			if !complexClose(got, tc.input, tc.epsilon) {
				t.Fatalf("expected %v, got %v", tc.input, got)
			}
		})
	}
}

func TestStandardLibraryInverseFunctionBranchSensitiveCases(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name    string
		arg     complex128
		want    complex128
		epsilon float64
	}{
		{name: "log", arg: complex(-1, 0), want: cmplx.Log(complex(-1, 0)), epsilon: 1e-10},
		{name: "sqrt", arg: complex(-1, 0), want: cmplx.Sqrt(complex(-1, 0)), epsilon: 1e-8},
		{name: "acosh", arg: complex(0.5, 0), want: cmplx.Acosh(complex(0.5, 0)), epsilon: 1e-7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := registry.Expand(tc.name, ast.Variable{Name: "x"})
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			got, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{"x": tc.arg})
			if err != nil {
				t.Fatalf("EvaluateMap returned error: %v", err)
			}
			if !complexClose(got, tc.want, tc.epsilon) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardLibraryRepresentativeHighPrecisionValidation(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name    string
		arg     complex128
		want    complex128
		epsilon float64
	}{
		{name: "mul", arg: complex(2.5, 0.25), want: complex(2.5, 0.25) * complex(0.5, -1.5), epsilon: 1e-8},
		{name: "sigmoid", arg: complex(0.75, -0.25), want: 1 / (1 + cmplx.Exp(-complex(0.75, -0.25))), epsilon: 1e-8},
		{name: "log", arg: complex(2.5, 0.5), want: cmplx.Log(complex(2.5, 0.5)), epsilon: 1e-8},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				expr ast.Expr
				err  error
			)
			switch tc.name {
			case "mul":
				expr, err = registry.Expand("mul", ast.Variable{Name: "x"}, ast.Variable{Name: "y"})
			default:
				expr, err = registry.Expand(tc.name, ast.Variable{Name: "x"})
			}
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}
			bindings := eval.MapBindings[eval.HighPrecisionComplex]{
				"x": eval.NewHighPrecisionComplex(256, real(tc.arg), imag(tc.arg)),
			}
			if tc.name == "mul" {
				bindings["y"] = eval.NewHighPrecisionComplex(256, 0.5, -1.5)
			}
			got, err := eval.Evaluate(expr, eval.NewHighPrecisionComplexBackend(eval.Precision{
				WorkingBits: 256,
				LogBranch:   eval.PrincipalLogBranch,
			}), bindings)
			if err != nil {
				t.Fatalf("Evaluate returned error: %v", err)
			}
			if !highPrecisionClose(got, tc.want, tc.epsilon) {
				t.Fatalf("unexpected high-precision result: %s", got.String())
			}
		})
	}
}

func TestStandardLibraryNormalizationRegression(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name      string
		maxNodes  int
		maxDepth  int
		argNeeded bool
	}{
		{name: "id", maxNodes: 1, maxDepth: 1, argNeeded: true},
		{name: "sigmoid", maxNodes: 150, maxDepth: 30, argNeeded: true},
		{name: "atanh", maxNodes: 250, maxDepth: 35, argNeeded: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				expr ast.Expr
				err  error
			)
			if tc.argNeeded {
				expr, err = registry.Expand(tc.name, ast.Variable{Name: "x"})
			} else {
				expr, err = registry.Expand(tc.name)
			}
			if err != nil {
				t.Fatalf("Expand returned error: %v", err)
			}

			normalized := normalize.Expr(expr)
			stats := StatsForExpr(normalized)
			if stats.NodeCount > tc.maxNodes {
				t.Fatalf("expected normalized node count <= %d, got %d", tc.maxNodes, stats.NodeCount)
			}
			if stats.TreeDepth > tc.maxDepth {
				t.Fatalf("expected normalized depth <= %d, got %d", tc.maxDepth, stats.TreeDepth)
			}
		})
	}
}

func TestStandardLibraryExpansionSizeChecks(t *testing.T) {
	registry := StandardLibrary()

	cases := []struct {
		name        string
		minNodes    int
		minDepth    int
		dependencyN int
	}{
		{name: "tan", minNodes: 1000, minDepth: 70, dependencyN: 10},
		{name: "atanh", minNodes: 150, minDepth: 20, dependencyN: 8},
		{name: "sigmoid", minNodes: 90, minDepth: 15, dependencyN: 6},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stats, err := registry.Stats(tc.name)
			if err != nil {
				t.Fatalf("Stats returned error: %v", err)
			}
			if stats.NodeCount < tc.minNodes {
				t.Fatalf("expected node count >= %d, got %d", tc.minNodes, stats.NodeCount)
			}
			if stats.TreeDepth < tc.minDepth {
				t.Fatalf("expected depth >= %d, got %d", tc.minDepth, stats.TreeDepth)
			}
			if stats.TransitiveDepCount < tc.dependencyN {
				t.Fatalf("expected transitive deps >= %d, got %d", tc.dependencyN, stats.TransitiveDepCount)
			}
		})
	}
}

func complexClose(got, want complex128, epsilon float64) bool {
	return math.Abs(real(got)-real(want)) <= epsilon && math.Abs(imag(got)-imag(want)) <= epsilon
}

func highPrecisionClose(got eval.HighPrecisionComplex, want complex128, epsilon float64) bool {
	gotRe, _ := got.Real().Float64()
	gotIm, _ := got.Imag().Float64()
	return math.Abs(gotRe-real(want)) <= epsilon && math.Abs(gotIm-imag(want)) <= epsilon
}
