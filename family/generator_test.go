package family

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"eml-parser/concepts"
)

func TestGenerateArtifactRawSeed(t *testing.T) {
	artifact, err := GenerateArtifact(
		Seed{FamilyID: 0, FamilyName: "raw_identity", SourceKind: SourceKindRaw, SourceRef: "x"},
		concepts.StandardLibrary(),
		SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 4, SampleCount: 3, Seed: 7},
		4,
	)
	if err != nil {
		t.Fatalf("GenerateArtifact returned error: %v", err)
	}
	if artifact.CanonicalKey != "x" {
		t.Fatalf("unexpected canonical key: %q", artifact.CanonicalKey)
	}
	if len(artifact.Samples) != 3 {
		t.Fatalf("unexpected sample count: %d", len(artifact.Samples))
	}
	if len(artifact.Samples[0].Points) != 4 {
		t.Fatalf("unexpected point count: %d", len(artifact.Samples[0].Points))
	}
	if artifact.Concept != nil {
		t.Fatalf("raw seed should not have concept provenance: %+v", artifact.Concept)
	}
}

func TestGenerateArtifactConceptSeed(t *testing.T) {
	artifact, err := GenerateArtifact(
		Seed{FamilyID: 1, FamilyName: "concept_exp", SourceKind: SourceKindConcept, SourceRef: "exp"},
		concepts.StandardLibrary(),
		SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 4, SampleCount: 2, Seed: 11},
		4,
	)
	if err != nil {
		t.Fatalf("GenerateArtifact returned error: %v", err)
	}
	if artifact.CanonicalKey != "eml(x, 1)" {
		t.Fatalf("unexpected canonical key: %q", artifact.CanonicalKey)
	}
	if artifact.Concept == nil || artifact.Concept.Concept != "exp" {
		t.Fatalf("missing concept provenance: %+v", artifact.Concept)
	}
}

func TestGenerateArtifactDeterministic(t *testing.T) {
	seed := Seed{FamilyID: 3, FamilyName: "concept_sigmoid", SourceKind: SourceKindConcept, SourceRef: "sigmoid"}
	sampling := SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 5, SampleCount: 3, Seed: 19}

	first, err := GenerateArtifact(seed, concepts.StandardLibrary(), sampling, 4)
	if err != nil {
		t.Fatalf("GenerateArtifact first run error: %v", err)
	}
	second, err := GenerateArtifact(seed, concepts.StandardLibrary(), sampling, 4)
	if err != nil {
		t.Fatalf("GenerateArtifact second run error: %v", err)
	}
	b1, _ := json.Marshal(first)
	b2, _ := json.Marshal(second)
	if string(b1) != string(b2) {
		t.Fatalf("artifacts differ between deterministic runs")
	}
}

func TestGenerateArtifactFailsOnNonReal(t *testing.T) {
	_, err := GenerateArtifact(
		Seed{FamilyID: 2, FamilyName: "concept_log", SourceKind: SourceKindConcept, SourceRef: "log"},
		concepts.StandardLibrary(),
		SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 4, SampleCount: 1, Seed: 0},
		4,
	)
	if err == nil {
		t.Fatalf("expected non-real generation failure")
	}
}

func TestWriteCuratedArtifacts(t *testing.T) {
	root := t.TempDir()
	paths, artifacts, err := WriteCuratedArtifacts(root)
	if err != nil {
		t.Fatalf("WriteCuratedArtifacts returned error: %v", err)
	}
	if len(paths) != len(CuratedSeeds()) || len(artifacts) != len(CuratedSeeds()) {
		t.Fatalf("unexpected artifact counts: %d paths %d artifacts", len(paths), len(artifacts))
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if len(data) == 0 {
			t.Fatalf("artifact file %q is empty", path)
		}
		if filepath.Dir(path) != filepath.Join(root, "artifacts", "equivalence") {
			t.Fatalf("artifact written to unexpected directory: %q", path)
		}
	}
}
