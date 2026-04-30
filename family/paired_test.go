package family

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"eml-parser/concepts"
)

func TestGeneratePairedDataset(t *testing.T) {
	artifact, err := GeneratePairedDataset(
		CuratedEquivalenceFamilies()[2],
		concepts.StandardLibrary(),
		pairedSamplingDomains(SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 4, SampleCount: 2, Seed: 3}),
		len(CuratedEquivalenceFamilies()),
	)
	if err != nil {
		t.Fatalf("GeneratePairedDataset returned error: %v", err)
	}
	if artifact.FormatVersion != PairedDatasetFormatVersion {
		t.Fatalf("unexpected format version: %q", artifact.FormatVersion)
	}
	if len(artifact.SamplingDomains) != 2 {
		t.Fatalf("unexpected domain count: %d", len(artifact.SamplingDomains))
	}
	if len(artifact.Groups) != 4 {
		t.Fatalf("unexpected group count: %d", len(artifact.Groups))
	}
	first := artifact.Groups[0]
	if first.DomainID == "" || first.GroupID == "" {
		t.Fatalf("missing group identifiers: %+v", first)
	}
	if len(first.Members) != 2 {
		t.Fatalf("unexpected member count: %d", len(first.Members))
	}
	if !first.Members[0].IsAnchor {
		t.Fatalf("first member should be anchor: %+v", first.Members[0])
	}
	if first.Members[1].RelationType != RelationKnownConceptEquivalence {
		t.Fatalf("unexpected member relation: %q", first.Members[1].RelationType)
	}
	if len(first.Points) != 4 {
		t.Fatalf("unexpected point count: %d", len(first.Points))
	}
	if len(first.Oracle) != len(CuratedEquivalenceFamilies()) {
		t.Fatalf("unexpected oracle width: %d", len(first.Oracle))
	}
}

func TestGeneratePairedDatasetDeterministic(t *testing.T) {
	spec := CuratedEquivalenceFamilies()[1]
	domains := pairedSamplingDomains(SamplingSpec{Variable: "x", Start: -0.75, Stop: 0.75, PointCount: 5, SampleCount: 2, Seed: 12})

	first, err := GeneratePairedDataset(spec, concepts.StandardLibrary(), domains, len(CuratedEquivalenceFamilies()))
	if err != nil {
		t.Fatalf("first GeneratePairedDataset error: %v", err)
	}
	second, err := GeneratePairedDataset(spec, concepts.StandardLibrary(), domains, len(CuratedEquivalenceFamilies()))
	if err != nil {
		t.Fatalf("second GeneratePairedDataset error: %v", err)
	}
	b1, _ := json.Marshal(first)
	b2, _ := json.Marshal(second)
	if string(b1) != string(b2) {
		t.Fatalf("paired dataset artifacts differ between deterministic runs")
	}
}

func TestGeneratePairedDatasetRejectsInvalidMemberOnAnyDomain(t *testing.T) {
	_, err := GeneratePairedDataset(
		EquivalenceFamilySpec{
			FamilyID:   7,
			FamilyName: "bad_paired_family",
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
		pairedSamplingDomains(DefaultSampling()),
		len(CuratedEquivalenceFamilies()),
	)
	if err == nil {
		t.Fatalf("expected invalid member failure")
	}
}

func TestWriteCuratedPairedDatasets(t *testing.T) {
	root := t.TempDir()
	paths, artifacts, err := WriteCuratedPairedDatasets(root)
	if err != nil {
		t.Fatalf("WriteCuratedPairedDatasets returned error: %v", err)
	}
	if len(paths) != len(CuratedEquivalenceFamilies()) || len(artifacts) != len(CuratedEquivalenceFamilies()) {
		t.Fatalf("unexpected artifact counts: %d paths %d artifacts", len(paths), len(artifacts))
	}
	for _, path := range paths {
		if filepath.Dir(path) != filepath.Join(root, "artifacts", "equivalence", "paired") {
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
