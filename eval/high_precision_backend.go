package eval

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/mshafiee/bigmath"
)

var (
	// ErrUnsupportedLogBranch marks a precision configuration that the current
	// backend does not implement.
	ErrUnsupportedLogBranch = errors.New("unsupported logarithm branch")

	// ErrLogZero is returned for log(0), which is undefined in the complex plane.
	ErrLogZero = errors.New("logarithm undefined for zero")
)

// HighPrecisionComplex is the concrete placeholder value type for future
// trusted evaluation backends.
//
// It stores real and imaginary parts separately as big.Float values and carries
// the working precision used to construct the value. This avoids coupling the
// rest of the evaluator to a specific external numeric library while still
// giving the codebase a concrete type to integrate with.
type HighPrecisionComplex struct {
	re   *big.Float
	im   *big.Float
	bits uint
}

// NewHighPrecisionComplex constructs a complex value at the requested working
// precision.
func NewHighPrecisionComplex(bits uint, re, im float64) HighPrecisionComplex {
	return HighPrecisionComplex{
		re:   new(big.Float).SetPrec(bits).SetFloat64(re),
		im:   new(big.Float).SetPrec(bits).SetFloat64(im),
		bits: bits,
	}
}

// NewHighPrecisionReal constructs a purely real value at the requested working
// precision.
func NewHighPrecisionReal(bits uint, value float64) HighPrecisionComplex {
	return NewHighPrecisionComplex(bits, value, 0)
}

// PrecisionBits reports the working precision associated with the value.
func (v HighPrecisionComplex) PrecisionBits() uint {
	return v.bits
}

// String returns a readable representation suitable for debugging and tests.
func (v HighPrecisionComplex) String() string {
	if v.re == nil || v.im == nil {
		return "<nil>"
	}
	return fmt.Sprintf("(%s + %si)", v.re.Text('g', -1), v.im.Text('g', -1))
}

// Real returns a defensive copy of the real component.
func (v HighPrecisionComplex) Real() *big.Float {
	if v.re == nil {
		return nil
	}
	return new(big.Float).SetPrec(v.bits).Set(v.re)
}

// Imag returns a defensive copy of the imaginary component.
func (v HighPrecisionComplex) Imag() *big.Float {
	if v.im == nil {
		return nil
	}
	return new(big.Float).SetPrec(v.bits).Set(v.im)
}

// HighPrecisionComplexBackend is the concrete backend shell for future trusted
// evaluation. It owns precision and branch policy and currently implements the
// paper-grounded EML operator via pure-Go bigmath real transcendentals.
type HighPrecisionComplexBackend struct {
	precision Precision
}

// NewHighPrecisionComplexBackend configures a future trusted backend boundary.
func NewHighPrecisionComplexBackend(precision Precision) HighPrecisionComplexBackend {
	if precision.WorkingBits == 0 {
		precision.WorkingBits = 256
	}
	if precision.LogBranch == "" {
		precision.LogBranch = PrincipalLogBranch
	}
	return HighPrecisionComplexBackend{precision: precision}
}

// Precision reports the configured evaluation policy.
func (b HighPrecisionComplexBackend) Precision() Precision {
	return b.precision
}

// One returns the distinguished constant terminal at backend precision.
func (b HighPrecisionComplexBackend) One() HighPrecisionComplex {
	return NewHighPrecisionReal(b.precision.WorkingBits, 1)
}

func (b HighPrecisionComplexBackend) EML(left, right HighPrecisionComplex) (HighPrecisionComplex, error) {
	if b.precision.LogBranch != PrincipalLogBranch {
		return HighPrecisionComplex{}, fmt.Errorf("%w: %q", ErrUnsupportedLogBranch, b.precision.LogBranch)
	}

	expLeft := b.exp(left)
	logRight, err := b.log(right)
	if err != nil {
		return HighPrecisionComplex{}, err
	}
	return newHighPrecisionComplexFromFloats(
		b.precision.WorkingBits,
		subFloat(expLeft.re, logRight.re, b.precision.WorkingBits),
		subFloat(expLeft.im, logRight.im, b.precision.WorkingBits),
	), nil
}

func (b HighPrecisionComplexBackend) exp(v HighPrecisionComplex) HighPrecisionComplex {
	prec := b.precision.WorkingBits
	expRe := bigmath.BigExp(v.re, prec)
	cosIm := bigmath.BigCos(v.im, prec)
	sinIm := bigmath.BigSin(v.im, prec)

	return newHighPrecisionComplexFromFloats(
		prec,
		mulFloat(expRe, cosIm, prec),
		mulFloat(expRe, sinIm, prec),
	)
}

func (b HighPrecisionComplexBackend) log(v HighPrecisionComplex) (HighPrecisionComplex, error) {
	prec := b.precision.WorkingBits
	if isZero(v.re) && isZero(v.im) {
		return HighPrecisionComplex{}, ErrLogZero
	}

	re2 := mulFloat(v.re, v.re, prec)
	im2 := mulFloat(v.im, v.im, prec)
	modulusSquared := addFloat(re2, im2, prec)
	logModulusSquared := bigmath.BigLog(modulusSquared, prec)
	half := new(big.Float).SetPrec(prec).SetFloat64(0.5)
	realPart := mulFloat(logModulusSquared, half, prec)
	imagPart := bigmath.BigAtan2(v.im, v.re, prec)

	return newHighPrecisionComplexFromFloats(prec, realPart, imagPart), nil
}

func newHighPrecisionComplexFromFloats(bits uint, re, im *big.Float) HighPrecisionComplex {
	return HighPrecisionComplex{
		re:   cloneFloat(re, bits),
		im:   cloneFloat(im, bits),
		bits: bits,
	}
}

func cloneFloat(x *big.Float, bits uint) *big.Float {
	if x == nil {
		return nil
	}
	return new(big.Float).SetPrec(bits).Set(x)
}

func addFloat(a, b *big.Float, bits uint) *big.Float {
	return new(big.Float).SetPrec(bits).Add(a, b)
}

func subFloat(a, b *big.Float, bits uint) *big.Float {
	return new(big.Float).SetPrec(bits).Sub(a, b)
}

func mulFloat(a, b *big.Float, bits uint) *big.Float {
	return new(big.Float).SetPrec(bits).Mul(a, b)
}

func isZero(x *big.Float) bool {
	return x != nil && x.Sign() == 0
}
