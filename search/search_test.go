package search

import (
	"testing"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/eval"
)

func TestNewCandidateNormalizesAndKeys(t *testing.T) {
	registry := concepts.StandardLibrary()
	expr, err := registry.ExpandSymbolic("id")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}

	candidate := NewCandidate(expr)
	if candidate.Key != "x" {
		t.Fatalf("expected canonical key x, got %q", candidate.Key)
	}
	if candidate.Stats.NodeCount != 1 || candidate.Stats.TreeDepth != 1 || candidate.Stats.LeafCount != 1 {
		t.Fatalf("unexpected stats: %+v", candidate.Stats)
	}
}

func TestCanonicalKeyDistinguishesDifferentTrees(t *testing.T) {
	a := NewCandidate(ast.Variable{Name: "x"})
	b := NewCandidate(ast.Variable{Name: "y"})
	if a.Key == b.Key {
		t.Fatalf("expected distinct keys, got %q", a.Key)
	}
}

func TestSubtreesIncludesRootAndChildren(t *testing.T) {
	expr := ast.Apply{
		Left: ast.Variable{Name: "x"},
		Right: ast.Apply{
			Left:  ast.One{},
			Right: ast.Variable{Name: "y"},
		},
	}

	subtrees := Subtrees(expr)
	if len(subtrees) != 5 {
		t.Fatalf("expected 5 subtrees, got %d", len(subtrees))
	}
	if subtrees[0].String() != "eml(x, eml(1, y))" {
		t.Fatalf("unexpected root subtree: %s", subtrees[0])
	}
}

func TestComplexMSEExactMatch(t *testing.T) {
	registry := concepts.StandardLibrary()
	expr, err := registry.ExpandSymbolic("exp")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}
	candidate := NewCandidate(expr)

	samples := []Sample[complex128]{
		{
			Vars:   map[string]complex128{"x": complex(0, 0)},
			Target: complex(1, 0),
		},
		{
			Vars:   map[string]complex128{"x": complex(1, 0)},
			Target: complex(2.718281828459045, 0),
		},
	}

	mse, err := ComplexMSE(candidate, eval.Complex128Backend{}, samples)
	if err != nil {
		t.Fatalf("ComplexMSE returned error: %v", err)
	}
	if mse > 1e-12 {
		t.Fatalf("expected near-zero mse, got %g", mse)
	}
}

func TestRealMSEPropagatesEvaluationError(t *testing.T) {
	candidate := NewCandidate(ast.Variable{Name: "x"})
	_, err := RealMSE(candidate, eval.Complex128Backend{}, []Sample[float64]{
		{
			Vars:   map[string]float64{},
			Target: 1,
		},
	})
	if err == nil {
		t.Fatal("expected evaluation error")
	}
}
