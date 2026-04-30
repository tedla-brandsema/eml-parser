package family

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"eml-parser/concepts"
)

func TestGenerateEquivalenceFamily(t *testing.T) {
	artifact, err := GenerateEquivalenceFamily(
		EquivalenceFamilySpec{
			FamilyID:   2,
			FamilyName: "exp_known_equivalence",
			Anchor: EntrySeed{
				Name:       "concept_exp_anchor",
				SourceKind: SourceKindConcept,
				SourceRef:  "exp",
			},
			Members: []MemberSpec{
				{
					Name:         "raw_exp_member",
					RelationType: RelationKnownConceptEquivalence,
					SourceKind:   SourceKindRaw,
					SourceRef:    "eml(x, 1)",
				},
			},
		},
		concepts.StandardLibrary(),
		SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 4, SampleCount: 2, Seed: 5},
	)
	if err != nil {
		t.Fatalf("GenerateEquivalenceFamily returned error: %v", err)
	}
	if artifact.FormatVersion != EquivalenceFamilyFormatVersion {
		t.Fatalf("unexpected format version: %q", artifact.FormatVersion)
	}
	if artifact.Anchor.CanonicalKey != "eml(x, 1)" {
		t.Fatalf("unexpected anchor canonical key: %q", artifact.Anchor.CanonicalKey)
	}
	if len(artifact.Members) != 1 {
		t.Fatalf("unexpected member count: %d", len(artifact.Members))
	}
	if artifact.Members[0].RelationType != RelationKnownConceptEquivalence {
		t.Fatalf("unexpected relation type: %q", artifact.Members[0].RelationType)
	}
	if artifact.Members[0].Concept != nil {
		t.Fatalf("raw member should not carry concept provenance: %+v", artifact.Members[0].Concept)
	}
	if artifact.Anchor.Concept == nil || artifact.Anchor.Concept.Concept != "exp" {
		t.Fatalf("missing anchor concept provenance: %+v", artifact.Anchor.Concept)
	}
	if len(artifact.SharedSamples) != 2 {
		t.Fatalf("unexpected shared sample count: %d", len(artifact.SharedSamples))
	}
	if len(artifact.SharedSamples[0].Points) != 4 {
		t.Fatalf("unexpected point count: %d", len(artifact.SharedSamples[0].Points))
	}
}

func TestGenerateEquivalenceFamilyDeterministic(t *testing.T) {
	spec := CuratedEquivalenceFamilies()[1]
	sampling := SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 5, SampleCount: 2, Seed: 9}

	first, err := GenerateEquivalenceFamily(spec, concepts.StandardLibrary(), sampling)
	if err != nil {
		t.Fatalf("first GenerateEquivalenceFamily error: %v", err)
	}
	second, err := GenerateEquivalenceFamily(spec, concepts.StandardLibrary(), sampling)
	if err != nil {
		t.Fatalf("second GenerateEquivalenceFamily error: %v", err)
	}
	b1, _ := json.Marshal(first)
	b2, _ := json.Marshal(second)
	if string(b1) != string(b2) {
		t.Fatalf("equivalence family artifacts differ between deterministic runs")
	}
}

func TestGenerateEquivalenceFamilyRejectsInvalidRelation(t *testing.T) {
	_, err := GenerateEquivalenceFamily(
		EquivalenceFamilySpec{
			FamilyID:   7,
			FamilyName: "invalid_relation",
			Anchor: EntrySeed{
				Name:       "anchor",
				SourceKind: SourceKindRaw,
				SourceRef:  "x",
			},
			Members: []MemberSpec{
				{
					Name:         "member",
					RelationType: RelationType("bogus"),
					SourceKind:   SourceKindRaw,
					SourceRef:    "x",
				},
			},
		},
		concepts.StandardLibrary(),
		DefaultSampling(),
	)
	if err == nil {
		t.Fatalf("expected invalid relation type failure")
	}
}

func TestGenerateEquivalenceFamilyRejectsDisagreement(t *testing.T) {
	_, err := GenerateEquivalenceFamily(
		EquivalenceFamilySpec{
			FamilyID:   8,
			FamilyName: "bad_equivalence",
			Anchor: EntrySeed{
				Name:       "anchor",
				SourceKind: SourceKindRaw,
				SourceRef:  "eml(x, 1)",
			},
			Members: []MemberSpec{
				{
					Name:         "wrong_member",
					RelationType: RelationKnownConceptEquivalence,
					SourceKind:   SourceKindRaw,
					SourceRef:    "x",
				},
			},
		},
		concepts.StandardLibrary(),
		SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 4, SampleCount: 1, Seed: 1},
	)
	if err == nil {
		t.Fatalf("expected numeric disagreement failure")
	}
}

func TestWriteCuratedEquivalenceFamilies(t *testing.T) {
	root := t.TempDir()
	paths, artifacts, err := WriteCuratedEquivalenceFamilies(root)
	if err != nil {
		t.Fatalf("WriteCuratedEquivalenceFamilies returned error: %v", err)
	}
	if len(paths) != len(CuratedEquivalenceFamilies()) || len(artifacts) != len(CuratedEquivalenceFamilies()) {
		t.Fatalf("unexpected artifact counts: %d paths %d artifacts", len(paths), len(artifacts))
	}
	for _, path := range paths {
		if filepath.Ext(path) != ".json" {
			t.Fatalf("unexpected artifact extension: %q", path)
		}
		if filepath.Dir(path) != filepath.Join(root, "artifacts", "equivalence") {
			t.Fatalf("artifact written to unexpected directory: %q", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		if len(data) == 0 {
			t.Fatalf("artifact file %q is empty", path)
		}
	}
}
