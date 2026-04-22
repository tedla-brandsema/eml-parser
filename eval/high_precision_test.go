package eval

import (
	"errors"
	"math"
	"math/cmplx"
	"testing"
)

type fakeHPValue struct {
	text string
	bits uint
}

func (v fakeHPValue) String() string     { return v.text }
func (v fakeHPValue) PrecisionBits() uint { return v.bits }

type fakeHPBackend struct{}

func (fakeHPBackend) One() fakeHPValue {
	return fakeHPValue{text: "1", bits: 256}
}

func (fakeHPBackend) EML(left, right fakeHPValue) (fakeHPValue, error) {
	return fakeHPValue{text: "eml(" + left.text + "," + right.text + ")", bits: 256}, nil
}

func (fakeHPBackend) Precision() Precision {
	return Precision{
		WorkingBits: 256,
		LogBranch:   PrincipalLogBranch,
	}
}

func TestFakeBackendSatisfiesHighPrecisionBoundary(t *testing.T) {
	var backend HighPrecisionBackend[fakeHPValue] = fakeHPBackend{}
	precision := backend.Precision()

	if precision.WorkingBits != 256 {
		t.Fatalf("expected 256 working bits, got %d", precision.WorkingBits)
	}
	if precision.LogBranch != PrincipalLogBranch {
		t.Fatalf("expected principal branch, got %q", precision.LogBranch)
	}
}

func TestHighPrecisionComplexBackendDefaults(t *testing.T) {
	backend := NewHighPrecisionComplexBackend(Precision{})
	precision := backend.Precision()

	if precision.WorkingBits != 256 {
		t.Fatalf("expected default precision of 256 bits, got %d", precision.WorkingBits)
	}
	if precision.LogBranch != PrincipalLogBranch {
		t.Fatalf("expected principal branch, got %q", precision.LogBranch)
	}
}

func TestHighPrecisionComplexOne(t *testing.T) {
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 192,
		LogBranch:   PrincipalLogBranch,
	})
	one := backend.One()

	if one.PrecisionBits() != 192 {
		t.Fatalf("expected value precision 192, got %d", one.PrecisionBits())
	}
	if got := one.String(); got != "(1 + 0i)" {
		t.Fatalf("unexpected one string: %s", got)
	}
}

func TestHighPrecisionComplexBackendEMLNotImplemented(t *testing.T) {
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 128,
		LogBranch:   PrincipalLogBranch,
	})

	got, err := backend.EML(NewHighPrecisionReal(128, 1), NewHighPrecisionReal(128, 1))
	if err != nil {
		t.Fatalf("expected working EML implementation, got %v", err)
	}

	want := cmplx.Exp(complex(1, 0)) - cmplx.Log(complex(1, 0))
	if !highPrecisionComplexClose(got, want, 1e-12) {
		t.Fatalf("expected %v, got %s", want, got.String())
	}
}

func TestEvaluateHighPrecisionOne(t *testing.T) {
	expr := mustParseTestExpr(t, "1")
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 160,
		LogBranch:   PrincipalLogBranch,
	})

	got, err := Evaluate(expr, backend, nil)
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if got.PrecisionBits() != 160 {
		t.Fatalf("expected precision 160, got %d", got.PrecisionBits())
	}
	if got.String() != "(1 + 0i)" {
		t.Fatalf("unexpected one value: %s", got.String())
	}
}

func TestEvaluateHighPrecisionVariableBinding(t *testing.T) {
	expr := mustParseTestExpr(t, "x")
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 224,
		LogBranch:   PrincipalLogBranch,
	})
	value := NewHighPrecisionComplex(224, 2, -3)

	got, err := Evaluate(expr, backend, MapBindings[HighPrecisionComplex]{
		"x": value,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if got.PrecisionBits() != 224 {
		t.Fatalf("expected precision 224, got %d", got.PrecisionBits())
	}
	if got.String() != value.String() {
		t.Fatalf("expected %s, got %s", value.String(), got.String())
	}
}

func TestEvaluateHighPrecisionEMLMatchesComplex128Reference(t *testing.T) {
	expr := mustParseTestExpr(t, "eml(1, x)")
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 256,
		LogBranch:   PrincipalLogBranch,
	})
	x := NewHighPrecisionComplex(256, 2.5, 0.5)

	got, err := Evaluate(expr, backend, MapBindings[HighPrecisionComplex]{
		"x": x,
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	want := cmplx.Exp(complex(1, 0)) - cmplx.Log(complex(2.5, 0.5))
	if !highPrecisionComplexClose(got, want, 1e-10) {
		t.Fatalf("expected %v, got %s", want, got.String())
	}
}

func TestEvaluateHighPrecisionUsesPrincipalBranch(t *testing.T) {
	expr := mustParseTestExpr(t, "eml(1, x)")
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 256,
		LogBranch:   PrincipalLogBranch,
	})

	got, err := Evaluate(expr, backend, MapBindings[HighPrecisionComplex]{
		"x": NewHighPrecisionComplex(256, -1, 0),
	})
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}

	want := cmplx.Exp(complex(1, 0)) - cmplx.Log(complex(-1, 0))
	if !highPrecisionComplexClose(got, want, 1e-10) {
		t.Fatalf("expected %v, got %s", want, got.String())
	}
}

func TestHighPrecisionRejectsLogZero(t *testing.T) {
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 128,
		LogBranch:   PrincipalLogBranch,
	})

	_, err := backend.EML(NewHighPrecisionReal(128, 1), NewHighPrecisionReal(128, 0))
	if !errors.Is(err, ErrLogZero) {
		t.Fatalf("expected ErrLogZero, got %v", err)
	}
}

func TestHighPrecisionRejectsUnsupportedBranch(t *testing.T) {
	backend := NewHighPrecisionComplexBackend(Precision{
		WorkingBits: 128,
		LogBranch:   LogBranch("nonprincipal"),
	})

	_, err := backend.EML(NewHighPrecisionReal(128, 1), NewHighPrecisionReal(128, 1))
	if !errors.Is(err, ErrUnsupportedLogBranch) {
		t.Fatalf("expected ErrUnsupportedLogBranch, got %v", err)
	}
}

func highPrecisionComplexClose(got HighPrecisionComplex, want complex128, epsilon float64) bool {
	gotRe, _ := got.Real().Float64()
	gotIm, _ := got.Imag().Float64()
	return math.Abs(gotRe-real(want)) <= epsilon && math.Abs(gotIm-imag(want)) <= epsilon
}
