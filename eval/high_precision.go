package eval

import "fmt"

// LogBranch identifies the chosen branch convention for complex logarithms.
type LogBranch string

const (
	// PrincipalLogBranch matches the paper's default complex-domain semantics.
	PrincipalLogBranch LogBranch = "principal"
)

// Precision captures the evaluation policy for a high-precision backend.
//
// WorkingBits is intentionally policy-level rather than implementation-level:
// concrete backends may internally use guard bits or wider temporaries.
type Precision struct {
	WorkingBits uint
	LogBranch   LogBranch
}

// HighPrecisionValue is the opaque numeric value type expected from an
// arbitrary-precision backend.
//
// The evaluator does not assume a specific implementation such as MPFR,
// big.Float pairs, or an external CAS bridge. It only requires the value to
// report the precision it is associated with and provide a human-readable form.
type HighPrecisionValue interface {
	fmt.Stringer
	PrecisionBits() uint
}

// HighPrecisionBackend is the boundary for future trusted evaluation backends.
//
// It extends the base EML backend with explicit precision and branch metadata.
// Concrete implementations remain free to choose their internal numeric engine.
type HighPrecisionBackend[V HighPrecisionValue] interface {
	Backend[V]
	Precision() Precision
}
