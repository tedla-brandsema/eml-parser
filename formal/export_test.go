package formal

import (
	"strings"
	"testing"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/search"
)

func TestExportExprNormalizes(t *testing.T) {
	registry := concepts.StandardLibrary()
	expr, err := registry.ExpandSymbolic("id")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}

	artifact := ExportExpr(expr)
	if artifact.Expression != "x" {
		t.Fatalf("expected normalized expression x, got %q", artifact.Expression)
	}
	if artifact.RootID != 1 {
		t.Fatalf("expected root id 1, got %d", artifact.RootID)
	}
	if len(artifact.Nodes) != 1 {
		t.Fatalf("expected one exported node, got %d", len(artifact.Nodes))
	}
	if artifact.Nodes[0].Kind != "var" || artifact.Nodes[0].Name != "x" {
		t.Fatalf("unexpected node: %+v", artifact.Nodes[0])
	}
	if artifact.Provenance != nil {
		t.Fatalf("expected nil provenance for raw expr export, got %+v", artifact.Provenance)
	}
}

func TestExportCandidateRetainsCandidateProvenance(t *testing.T) {
	candidate := search.NewCandidate(ast.Apply{
		Left:  ast.Variable{Name: "x"},
		Right: ast.One{},
	})

	artifact := ExportCandidate(candidate)
	if artifact.Provenance == nil {
		t.Fatal("expected provenance")
	}
	if artifact.Provenance.Source != "candidate" {
		t.Fatalf("unexpected provenance source: %+v", artifact.Provenance)
	}
	if artifact.Provenance.CandidateKey != candidate.Key {
		t.Fatalf("expected candidate key %q, got %+v", candidate.Key, artifact.Provenance)
	}
}

func TestExportConceptRetainsConceptProvenance(t *testing.T) {
	registry := concepts.StandardLibrary()
	artifact, err := ExportConcept(registry, "id")
	if err != nil {
		t.Fatalf("ExportConcept returned error: %v", err)
	}
	if artifact.Expression != "x" {
		t.Fatalf("expected normalized concept expression x, got %q", artifact.Expression)
	}
	if artifact.Provenance == nil || artifact.Provenance.Concept == nil {
		t.Fatalf("expected concept provenance, got %+v", artifact.Provenance)
	}
	if artifact.Provenance.Source != "concept" {
		t.Fatalf("unexpected provenance source: %+v", artifact.Provenance)
	}
	if artifact.Provenance.Concept.Name != "id" {
		t.Fatalf("unexpected concept provenance: %+v", artifact.Provenance.Concept)
	}
	if !strings.Contains(artifact.Provenance.Concept.Definition, "id(x) :=") {
		t.Fatalf("unexpected concept definition: %+v", artifact.Provenance.Concept)
	}
	if len(artifact.Provenance.Concept.DependencyPaths) == 0 {
		t.Fatalf("expected dependency paths, got %+v", artifact.Provenance.Concept)
	}
}

func TestArtifactJSON(t *testing.T) {
	artifact := ExportExpr(ast.Variable{Name: "x"})
	payload, err := artifact.JSON()
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}
	for _, expected := range []string{
		`"format_version": "eml-formal-v1"`,
		`"expression": "x"`,
		`"kind": "var"`,
		`"name": "x"`,
	} {
		if !strings.Contains(payload, expected) {
			t.Fatalf("expected %q in payload, got %q", expected, payload)
		}
	}
}
