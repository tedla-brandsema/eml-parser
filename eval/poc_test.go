package eval

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestPoCNestedTreeParsesAndEvaluates(t *testing.T) {
	expr := mustParseTestExpr(t, "eml(eml(1, x), eml(y, 1))")

	fast, err := EvaluateMap(expr, Complex128Backend{}, map[string]complex128{
		"x": complex(2, 0.5),
		"y": complex(-0.25, 1.25),
	})
	if err != nil {
		t.Fatalf("fast evaluation returned error: %v", err)
	}

	want := cmplx.Exp(cmplx.Exp(complex(1, 0))-cmplx.Log(complex(2, 0.5))) - cmplx.Log(cmplx.Exp(complex(-0.25, 1.25))-cmplx.Log(complex(1, 0)))
	if !almostEqual(fast, want, 1e-12) {
		t.Fatalf("expected %v, got %v", want, fast)
	}
}

func TestPoCOneNeutralizesLogTerm(t *testing.T) {
	expr := mustParseTestExpr(t, "eml(x, 1)")
	x := complex(0.75, -1.5)

	got, err := EvaluateMap(expr, Complex128Backend{}, map[string]complex128{
		"x": x,
	})
	if err != nil {
		t.Fatalf("EvaluateMap returned error: %v", err)
	}

	want := cmplx.Exp(x)
	if !almostEqual(got, want, 1e-12) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestPoCBackendsAgreeOnRepresentativeExpressions(t *testing.T) {
	cases := []struct {
		name  string
		input string
		vars  map[string]complex128
	}{
		{
			name:  "simple_log_branch",
			input: "eml(1, x)",
			vars: map[string]complex128{
				"x": complex(-1, 0),
			},
		},
		{
			name:  "mixed_nested_tree",
			input: "eml(eml(x, 1), eml(1, y))",
			vars: map[string]complex128{
				"x": complex(0.5, -0.75),
				"y": complex(2.25, 0.25),
			},
		},
		{
			name:  "nontrivial_complex_inputs",
			input: "eml(eml(1, x), y)",
			vars: map[string]complex128{
				"x": complex(1.5, 0.25),
				"y": complex(0.75, -0.5),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr := mustParseTestExpr(t, tc.input)

			fast, err := EvaluateMap(expr, Complex128Backend{}, tc.vars)
			if err != nil {
				t.Fatalf("fast evaluation returned error: %v", err)
			}

			highPrecisionVars := make(map[string]HighPrecisionComplex, len(tc.vars))
			for name, value := range tc.vars {
				highPrecisionVars[name] = NewHighPrecisionComplex(256, real(value), imag(value))
			}

			high, err := Evaluate(expr, NewHighPrecisionComplexBackend(Precision{
				WorkingBits: 256,
				LogBranch:   PrincipalLogBranch,
			}), MapBindings[HighPrecisionComplex](highPrecisionVars))
			if err != nil {
				t.Fatalf("high-precision evaluation returned error: %v", err)
			}

			if !highPrecisionComplexClose(high, fast, 1e-10) {
				t.Fatalf("backends diverged: fast=%v high=%s", fast, high.String())
			}
		})
	}
}

func TestPoCIncreasingPrecisionStabilizesResult(t *testing.T) {
	expr := mustParseTestExpr(t, "eml(eml(1, x), y)")
	vars := map[string]HighPrecisionComplex{
		"x": NewHighPrecisionComplex(128, 1.75, -0.25),
		"y": NewHighPrecisionComplex(128, -0.5, 0.75),
	}

	low, err := Evaluate(expr, NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 128,
		LogBranch:   PrincipalLogBranch,
	}), MapBindings[HighPrecisionComplex](vars))
	if err != nil {
		t.Fatalf("128-bit evaluation returned error: %v", err)
	}

	vars["x"] = NewHighPrecisionComplex(256, 1.75, -0.25)
	vars["y"] = NewHighPrecisionComplex(256, -0.5, 0.75)
	high, err := Evaluate(expr, NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 256,
		LogBranch:   PrincipalLogBranch,
	}), MapBindings[HighPrecisionComplex](vars))
	if err != nil {
		t.Fatalf("256-bit evaluation returned error: %v", err)
	}

	lowRe, _ := low.Real().Float64()
	lowIm, _ := low.Imag().Float64()
	highRe, _ := high.Real().Float64()
	highIm, _ := high.Imag().Float64()

	if math.Abs(lowRe-highRe) > 1e-8 || math.Abs(lowIm-highIm) > 1e-8 {
		t.Fatalf("expected 128-bit and 256-bit results to be close, got low=%s high=%s", low.String(), high.String())
	}
}
