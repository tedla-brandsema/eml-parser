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

func TestReplaceSubtree(t *testing.T) {
	expr := ast.Apply{
		Left:  ast.Variable{Name: "x"},
		Right: ast.Variable{Name: "y"},
	}

	got, err := ReplaceSubtree(expr, 1, ast.One{})
	if err != nil {
		t.Fatalf("ReplaceSubtree returned error: %v", err)
	}
	if got.String() != "eml(1, y)" {
		t.Fatalf("unexpected replacement result: %s", got)
	}
}

func TestMutateByReplacement(t *testing.T) {
	expr := ast.Apply{
		Left:  ast.Variable{Name: "x"},
		Right: ast.Variable{Name: "y"},
	}
	mutations := MutateByReplacement(expr, []ast.Expr{ast.One{}}, Bounds{
		MaxDepth: 3,
		MaxNodes: 5,
	})
	if len(mutations) != 3 {
		t.Fatalf("expected 3 unique mutations, got %d", len(mutations))
	}
}

func TestWithinBounds(t *testing.T) {
	expr := ast.Apply{
		Left: ast.Variable{Name: "x"},
		Right: ast.Apply{
			Left:  ast.One{},
			Right: ast.Variable{Name: "y"},
		},
	}
	if !WithinBounds(expr, Bounds{MaxDepth: 3, MaxNodes: 5}) {
		t.Fatal("expected expression to satisfy bounds")
	}
	if WithinBounds(expr, Bounds{MaxDepth: 2}) {
		t.Fatal("expected depth bound failure")
	}
}

func TestEnumerateBounded(t *testing.T) {
	exprs := EnumerateBounded(AtomicSeeds("x"), Bounds{
		MaxDepth: 2,
		MaxNodes: 3,
	})
	if len(exprs) < 3 {
		t.Fatalf("expected at least seeds plus one composite, got %d", len(exprs))
	}
}

func TestUniqueCandidatesDeduplicatesByNormalizedKey(t *testing.T) {
	registry := concepts.StandardLibrary()
	idExpr, err := registry.ExpandSymbolic("id")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}

	unique := UniqueCandidates([]ast.Expr{
		ast.Variable{Name: "x"},
		idExpr,
	})
	if len(unique) != 1 {
		t.Fatalf("expected 1 unique candidate, got %d", len(unique))
	}
	if unique[0].Key != "x" {
		t.Fatalf("unexpected canonical key: %q", unique[0].Key)
	}
}

func TestRealRangeSamples(t *testing.T) {
	samples := RealRangeSamples("x", -1, 1, 3, func(x float64) float64 { return x * x })
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(samples))
	}
	if samples[0].Vars["x"] != -1 || samples[2].Vars["x"] != 1 {
		t.Fatalf("unexpected sample coordinates: %+v", samples)
	}
	if samples[1].Target != 0 {
		t.Fatalf("unexpected midpoint target: %v", samples[1].Target)
	}
}

func TestComplexGridSamples(t *testing.T) {
	samples := ComplexGridSamples("z", []float64{0, 1}, []float64{-1, 1}, func(z complex128) complex128 { return z })
	if len(samples) != 4 {
		t.Fatalf("expected 4 samples, got %d", len(samples))
	}
}

func TestRealBenchmarkFixturesScoreExactly(t *testing.T) {
	fixtures, err := RealBenchmarkFixtures()
	if err != nil {
		t.Fatalf("RealBenchmarkFixtures returned error: %v", err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			candidate := NewCandidate(fixture.Expr)
			mse, err := RealMSE(candidate, eval.Complex128Backend{}, fixture.Samples)
			if err != nil {
				t.Fatalf("RealMSE returned error: %v", err)
			}
			if mse > 1e-12 {
				t.Fatalf("expected near-zero mse, got %g", mse)
			}
		})
	}
}

func TestComplexBenchmarkFixturesScoreExactly(t *testing.T) {
	fixtures, err := ComplexBenchmarkFixtures()
	if err != nil {
		t.Fatalf("ComplexBenchmarkFixtures returned error: %v", err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			candidate := NewCandidate(fixture.Expr)
			mse, err := ComplexMSE(candidate, eval.Complex128Backend{}, fixture.Samples)
			if err != nil {
				t.Fatalf("ComplexMSE returned error: %v", err)
			}
			if mse > 1e-12 {
				t.Fatalf("expected near-zero mse, got %g", mse)
			}
		})
	}
}

func TestRealBenchmarkFixtureByName(t *testing.T) {
	fixture, err := RealBenchmarkFixtureByName("exp_real_small")
	if err != nil {
		t.Fatalf("RealBenchmarkFixtureByName returned error: %v", err)
	}
	if fixture.Name != "exp_real_small" {
		t.Fatalf("unexpected fixture: %+v", fixture)
	}
}

func TestEnumerativeRealSearchFindsExactExpression(t *testing.T) {
	fixture, err := RealBenchmarkFixtureByName("exp_real_small")
	if err != nil {
		t.Fatalf("RealBenchmarkFixtureByName returned error: %v", err)
	}

	results, err := EnumerativeRealSearch(fixture, eval.Complex128Backend{}, SearchOptions{
		Bounds: Bounds{
			MaxDepth: 2,
			MaxNodes: 3,
		},
		TopN: 5,
	})
	if err != nil {
		t.Fatalf("EnumerativeRealSearch returned error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected non-empty search results")
	}
	if results[0].Candidate.Key != "eml(x, 1)" {
		t.Fatalf("expected exp candidate first, got %q", results[0].Candidate.Key)
	}
	if results[0].Score > 1e-12 {
		t.Fatalf("expected near-zero score, got %g", results[0].Score)
	}
}
