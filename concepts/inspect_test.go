package concepts

import "testing"

func TestDependencyPaths(t *testing.T) {
	registry := StandardLibrary()

	paths, err := registry.DependencyPaths("id")
	if err != nil {
		t.Fatalf("DependencyPaths returned error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 dependency paths, got %d", len(paths))
	}
	if got := joinPath(paths[0]); got != "id -> exp" {
		t.Fatalf("unexpected first path: %s", got)
	}
	if got := joinPath(paths[1]); got != "id -> log -> exp" {
		t.Fatalf("unexpected second path: %s", got)
	}
}

func TestInspectCombinesDefinitionExpansionAndNormalization(t *testing.T) {
	registry := StandardLibrary()

	inspection, err := registry.Inspect("id")
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}
	if inspection.Definition != "id(x) := exp(log(x))" {
		t.Fatalf("unexpected definition: %s", inspection.Definition)
	}
	if inspection.Expanded.String() != "eml(eml(1, eml(eml(1, x), 1)), 1)" {
		t.Fatalf("unexpected expanded form: %s", inspection.Expanded)
	}
	if inspection.Normalized.String() != "x" {
		t.Fatalf("unexpected normalized form: %s", inspection.Normalized)
	}
	if inspection.ExpandedNodeDelta <= 0 || inspection.ExpandedDepthDelta <= 0 {
		t.Fatalf("expected positive deltas, got nodes=%d depth=%d", inspection.ExpandedNodeDelta, inspection.ExpandedDepthDelta)
	}
}

func joinPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	out := path[0]
	for i := 1; i < len(path); i++ {
		out += " -> " + path[i]
	}
	return out
}
