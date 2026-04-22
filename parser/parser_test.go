package parser

import (
	"testing"

	"eml-parser/ast"
)

func TestParseStringOne(t *testing.T) {
	expr, err := ParseString("1")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	if _, ok := expr.(ast.One); !ok {
		t.Fatalf("expected ast.One, got %T", expr)
	}
}

func TestParseStringVariable(t *testing.T) {
	expr, err := ParseString("x")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	v, ok := expr.(ast.Variable)
	if !ok {
		t.Fatalf("expected ast.Variable, got %T", expr)
	}
	if v.Name != "x" {
		t.Fatalf("expected variable x, got %q", v.Name)
	}
}

func TestParseStringApply(t *testing.T) {
	expr, err := ParseString("eml(1, x)")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	app, ok := expr.(ast.Apply)
	if !ok {
		t.Fatalf("expected ast.Apply, got %T", expr)
	}
	if app.Left.String() != "1" {
		t.Fatalf("unexpected left operand: %s", app.Left)
	}
	if app.Right.String() != "x" {
		t.Fatalf("unexpected right operand: %s", app.Right)
	}
}

func TestParseStringNestedApply(t *testing.T) {
	expr, err := ParseString("eml(eml(1, x), y)")
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	app, ok := expr.(ast.Apply)
	if !ok {
		t.Fatalf("expected ast.Apply, got %T", expr)
	}
	if app.Right.String() != "y" {
		t.Fatalf("unexpected right operand: %s", app.Right)
	}
}

func TestParseRejectsOtherNumericLiterals(t *testing.T) {
	_, err := ParseString("2")
	if err == nil {
		t.Fatal("expected error for numeric literal 2")
	}
}
