package eval

import (
	"math/cmplx"
)

// Complex128Backend evaluates EML expressions using Go's complex128 support.
//
// This is a fast screening backend, not a high-confidence verification backend.
// It follows the paper's complex-domain semantics with the principal branch of
// the logarithm supplied by math/cmplx.
type Complex128Backend struct{}

func (Complex128Backend) One() complex128 {
	return complex(1, 0)
}

func (Complex128Backend) EML(left, right complex128) (complex128, error) {
	return cmplx.Exp(left) - cmplx.Log(right), nil
}
