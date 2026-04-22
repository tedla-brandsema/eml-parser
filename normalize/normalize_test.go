package normalize

import (
	"testing"

	"eml-parser/ast"
	"eml-parser/concepts"
)

func TestNormalizeCollapsesIdentityExpansion(t *testing.T) {
	registry := concepts.StandardLibrary()
	expr, err := registry.ExpandSymbolic("id")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}

	got := Expr(expr)
	if got.String() != "x" {
		t.Fatalf("expected x, got %s", got.String())
	}
}

func TestNormalizeCollapsesExpZeroToOne(t *testing.T) {
	expr := ast.Apply{
		Left: rawZero(),
		Right: ast.One{},
	}

	got := Expr(expr)
	if got.String() != "1" {
		t.Fatalf("expected 1, got %s", got.String())
	}
}

func TestNormalizeIsIdempotent(t *testing.T) {
	registry := concepts.StandardLibrary()
	expr, err := registry.ExpandSymbolic("tan")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}

	once := Expr(expr)
	twice := Expr(once)
	if once.String() != twice.String() {
		t.Fatalf("expected idempotent normalization, got %s then %s", once.String(), twice.String())
	}
}
