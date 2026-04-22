// Package eval evaluates EML ASTs using pluggable numeric backends.
//
// The initial implementation follows the EML paper's complex-domain reading of
// the operator as Exp-Minus-Log:
//
//	eml(a, b) = exp(a) - log(b)
//
// using the principal branch for the complex logarithm.
package eval
