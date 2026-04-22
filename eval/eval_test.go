package eval

import (
	"math"
	"math/cmplx"
	"testing"

	"eml-parser/ast"
	"eml-parser/parser"
)

func TestEvaluateOne(t *testing.T) {
	expr, err := parser.ParseString("1")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	got, err := Evaluate(expr, Complex128Backend{}, nil)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if got != complex(1, 0) {
		t.Fatalf("expected 1, got %v", got)
	}
}

func TestEvaluateVariable(t *testing.T) {
	expr, err := parser.ParseString("x")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	got, err := EvaluateMap(expr, Complex128Backend{}, map[string]complex128{
		"x": complex(2, -3),
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	if got != complex(2, -3) {
		t.Fatalf("expected bound variable value, got %v", got)
	}
}

func TestEvaluateEMLMatchesPaperOperator(t *testing.T) {
	expr, err := parser.ParseString("eml(1, x)")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	x := complex(2.5, 0.5)
	got, err := EvaluateMap(expr, Complex128Backend{}, map[string]complex128{
		"x": x,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	want := cmplx.Exp(complex(1, 0)) - cmplx.Log(x)
	if !almostEqual(got, want, 1e-12) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestEvaluateUsesPrincipalBranchAtNegativeRealAxis(t *testing.T) {
	expr, err := parser.ParseString("eml(1, x)")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	got, err := EvaluateMap(expr, Complex128Backend{}, map[string]complex128{
		"x": complex(-1, 0),
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	want := cmplx.Exp(complex(1, 0)) - cmplx.Log(complex(-1, 0))
	if !almostEqual(got, want, 1e-12) {
		t.Fatalf("expected principal-branch result %v, got %v", want, got)
	}
}

func TestEvaluateRejectsUnboundVariable(t *testing.T) {
	expr, err := parser.ParseString("x")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	_, err = Evaluate(expr, Complex128Backend{}, nil)
	if err == nil {
		t.Fatal("expected error for unbound variable")
	}
}

func TestEvaluateWithBindingsInterface(t *testing.T) {
	expr := mustParseTestExpr(t, "eml(x, 1)")
	bindings := MapBindings[complex128]{
		"x": complex(0, math.Pi),
	}
	got, err := Evaluate(expr, Complex128Backend{}, bindings)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	want := cmplx.Exp(complex(0, math.Pi)) - cmplx.Log(complex(1, 0))
	if !almostEqual(got, want, 1e-12) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func mustParseTestExpr(t *testing.T, input string) ast.Expr {
	t.Helper()

	expr, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}
	return expr
}

func almostEqual(a, b complex128, epsilon float64) bool {
	return math.Abs(real(a)-real(b)) <= epsilon && math.Abs(imag(a)-imag(b)) <= epsilon
}
