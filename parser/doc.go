// Package parser implements the initial EML language frontend.
//
// The current grammar is intentionally narrow and grounded in the EML paper's
// core representation of expressions as binary trees:
//
//	expr := "1"
//	      | identifier
//	      | "eml" "(" expr "," expr ")"
//
// This keeps the parser aligned with the paper's minimal basis: the
// distinguished constant 1, input variables, and repeated binary EML nodes.
package parser
