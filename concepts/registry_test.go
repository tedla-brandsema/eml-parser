package concepts

import (
	"errors"
	"testing"

	"eml-parser/ast"
	"eml-parser/eval"
)

func TestStandardLibraryExpExpandsToRawEML(t *testing.T) {
	registry := StandardLibrary()

	got, err := registry.Expand("exp", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	if got.String() != "eml(x, 1)" {
		t.Fatalf("expected eml(x, 1), got %s", got.String())
	}
}

func TestCompositeConceptExpandsRecursively(t *testing.T) {
	registry := StandardLibrary()
	if err := registry.Register(Definition{
		Name:   "doubleexp",
		Params: []string{"x"},
		Body:   Ref("exp", Ref("exp", P("x"))),
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	got, err := registry.Expand("doubleexp", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	if got.String() != "eml(eml(x, 1), 1)" {
		t.Fatalf("unexpected expansion: %s", got.String())
	}
}

func TestCompositeConceptSupportsPartialRawTrees(t *testing.T) {
	registry := StandardLibrary()
	if err := registry.Register(Definition{
		Name:   "exp_then_eml",
		Params: []string{"x", "y"},
		Body:   EML(Ref("exp", P("x")), P("y")),
	}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	got, err := registry.Expand("exp_then_eml", ast.Variable{Name: "x"}, ast.Variable{Name: "y"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	if got.String() != "eml(eml(x, 1), y)" {
		t.Fatalf("unexpected expansion: %s", got.String())
	}
}

func TestExpandDetectsCycle(t *testing.T) {
	registry := NewRegistry()
	mustRegisterTestConcept(t, registry, Definition{
		Name:   "a",
		Params: []string{"x"},
		Body:   Ref("b", P("x")),
	})
	mustRegisterTestConcept(t, registry, Definition{
		Name:   "b",
		Params: []string{"x"},
		Body:   Ref("a", P("x")),
	})

	_, err := registry.Expand("a", ast.Variable{Name: "x"})
	if !errors.Is(err, ErrConceptCycle) {
		t.Fatalf("expected ErrConceptCycle, got %v", err)
	}
}

func TestExpandRejectsUnknownConcept(t *testing.T) {
	registry := NewRegistry()
	_, err := registry.Expand("missing", ast.Variable{Name: "x"})
	if !errors.Is(err, ErrUnknownConcept) {
		t.Fatalf("expected ErrUnknownConcept, got %v", err)
	}
}

func TestNamesAreSorted(t *testing.T) {
	registry := NewRegistry()
	mustRegisterTestConcept(t, registry, Definition{Name: "z", Body: ConstOne()})
	mustRegisterTestConcept(t, registry, Definition{Name: "a", Body: ConstOne()})

	names := registry.Names()
	if len(names) != 2 || names[0] != "a" || names[1] != "z" {
		t.Fatalf("unexpected names: %v", names)
	}
}

func TestExpandSymbolicUsesParameterNames(t *testing.T) {
	registry := StandardLibrary()
	got, err := registry.ExpandSymbolic("exp")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}
	if got.String() != "eml(x, 1)" {
		t.Fatalf("unexpected symbolic expansion: %s", got.String())
	}
}

func TestDependencies(t *testing.T) {
	registry := StandardLibrary()

	direct, err := registry.DirectDependencies("tan")
	if err != nil {
		t.Fatalf("DirectDependencies returned error: %v", err)
	}
	if len(direct) != 3 || direct[0] != "cos" || direct[1] != "div" || direct[2] != "sin" {
		t.Fatalf("unexpected direct deps: %v", direct)
	}

	transitive, err := registry.TransitiveDependencies("tan")
	if err != nil {
		t.Fatalf("TransitiveDependencies returned error: %v", err)
	}
	if len(transitive) == 0 {
		t.Fatal("expected transitive deps for tan")
	}
}

func TestStats(t *testing.T) {
	registry := StandardLibrary()

	stats, err := registry.Stats("exp")
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if stats.Concept != "exp" {
		t.Fatalf("unexpected concept: %q", stats.Concept)
	}
	if stats.NodeCount != 3 || stats.TreeDepth != 2 || stats.LeafCount != 2 {
		t.Fatalf("unexpected tree stats: %+v", stats)
	}
	if stats.DirectDependencyCount != 0 || stats.TransitiveDepCount != 0 {
		t.Fatalf("unexpected dependency stats: %+v", stats)
	}
}

func TestExpandSymbolicCachesNamedExpansion(t *testing.T) {
	registry := StandardLibrary()

	if registry.symbolicCacheSize() != 0 {
		t.Fatalf("expected empty cache, got %d", registry.symbolicCacheSize())
	}

	first, err := registry.ExpandSymbolic("tan")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}
	if registry.symbolicCacheSize() != 1 {
		t.Fatalf("expected cache size 1, got %d", registry.symbolicCacheSize())
	}

	second, err := registry.ExpandSymbolic("tan")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}
	if registry.symbolicCacheSize() != 1 {
		t.Fatalf("expected cache size to stay 1, got %d", registry.symbolicCacheSize())
	}
	if first.String() != second.String() {
		t.Fatalf("expected identical symbolic expansions, got %s and %s", first.String(), second.String())
	}
}

func TestRegisterInvalidatesSymbolicCache(t *testing.T) {
	registry := StandardLibrary()

	if _, err := registry.ExpandSymbolic("exp"); err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}
	if registry.symbolicCacheSize() == 0 {
		t.Fatal("expected symbolic cache to be populated")
	}

	mustRegisterTestConcept(t, registry, Definition{
		Name:   "tripleexp",
		Params: []string{"x"},
		Body:   Ref("exp", Ref("exp", Ref("exp", P("x")))),
	})
	if registry.symbolicCacheSize() != 0 {
		t.Fatalf("expected cache invalidation on register, got size %d", registry.symbolicCacheSize())
	}
}

func TestExpandedConceptEvaluatesWithExistingBackend(t *testing.T) {
	registry := StandardLibrary()
	mustRegisterTestConcept(t, registry, Definition{
		Name:   "doubleexp",
		Params: []string{"x"},
		Body:   Ref("exp", Ref("exp", P("x"))),
	})

	expanded, err := registry.Expand("doubleexp", ast.Variable{Name: "x"})
	if err != nil {
		t.Fatalf("Expand returned error: %v", err)
	}

	got, err := eval.EvaluateMap(expanded, eval.Complex128Backend{}, map[string]complex128{
		"x": complex(0, 0),
	})
	if err != nil {
		t.Fatalf("EvaluateMap returned error: %v", err)
	}

	if real(got) <= 0 {
		t.Fatalf("expected positive real result, got %v", got)
	}
}

func mustRegisterTestConcept(t *testing.T, registry *Registry, def Definition) {
	t.Helper()
	if err := registry.Register(def); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
}
