// Package concepts provides a recursive dictionary of named mathematical
// constructions above the raw EML parser.
//
// The parser remains intentionally small and only understands atomic EML:
// the constant 1, variables, and binary eml(a, b) nodes. Higher-level
// mathematical concepts are represented here as named, parameterized
// definitions that expand recursively down to raw EML ASTs.
package concepts
