package family

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"eml-parser/concepts"
)

func TestGenerateSnippetDataset(t *testing.T) {
	artifact, err := GenerateSnippetDataset(
		SnippetTargetSpec{
			TargetID:   "raw_exp3",
			TargetName: "Raw Exp Depth 3",
			SourceKind: SourceKindRaw,
			SourceRef:  "eml(eml(eml(x, 1), 1), 1)",
			Selectors: []SnippetSelector{
				{Name: "exp2", PreorderIndex: 1},
				{Name: "exp1", PreorderIndex: 2},
				{Name: "x_leaf", PreorderIndex: 3},
			},
		},
		concepts.StandardLibrary(),
		snippetSamplingDomains(SamplingSpec{Variable: "x", Start: 0.25, Stop: 1.25, PointCount: 4, SampleCount: 2, Seed: 7}),
	)
	if err != nil {
		t.Fatalf("GenerateSnippetDataset returned error: %v", err)
	}
	if artifact.FormatVersion != SnippetDatasetFormatVersion {
		t.Fatalf("unexpected format version: %q", artifact.FormatVersion)
	}
	if len(artifact.Snippets) != 3 {
		t.Fatalf("unexpected snippet count: %d", len(artifact.Snippets))
	}
	if len(artifact.SamplingDomains) != 2 {
		t.Fatalf("unexpected domain count: %d", len(artifact.SamplingDomains))
	}
	if len(artifact.WholeSamples) != 4 {
		t.Fatalf("unexpected whole sample count: %d", len(artifact.WholeSamples))
	}
	if len(artifact.SnippetSamples) != 12 {
		t.Fatalf("unexpected snippet sample count: %d", len(artifact.SnippetSamples))
	}
	if artifact.Snippets[0].TreePath != "root.L" {
		t.Fatalf("unexpected snippet path: %q", artifact.Snippets[0].TreePath)
	}
	if artifact.Snippets[1].TreePath != "root.L.L" {
		t.Fatalf("unexpected nested snippet path: %q", artifact.Snippets[1].TreePath)
	}
}

func TestGenerateSnippetDatasetDeterministic(t *testing.T) {
	target := CuratedSnippetTargets()[0]
	domains := snippetSamplingDomains(SamplingSpec{Variable: "x", Start: 0.25, Stop: 1.25, PointCount: 5, SampleCount: 2, Seed: 5})
	first, err := GenerateSnippetDataset(target, concepts.StandardLibrary(), domains)
	if err != nil {
		t.Fatalf("first GenerateSnippetDataset error: %v", err)
	}
	second, err := GenerateSnippetDataset(target, concepts.StandardLibrary(), domains)
	if err != nil {
		t.Fatalf("second GenerateSnippetDataset error: %v", err)
	}
	b1, _ := json.Marshal(first)
	b2, _ := json.Marshal(second)
	if string(b1) != string(b2) {
		t.Fatalf("snippet datasets differ between deterministic runs")
	}
}

func TestGenerateSnippetDatasetRejectsInvalidSelector(t *testing.T) {
	_, err := GenerateSnippetDataset(
		SnippetTargetSpec{
			TargetID:   "bad_selector",
			TargetName: "Bad Selector",
			SourceKind: SourceKindRaw,
			SourceRef:  "eml(x, 1)",
			Selectors: []SnippetSelector{
				{Name: "missing", PreorderIndex: 99},
			},
		},
		concepts.StandardLibrary(),
		snippetSamplingDomains(DefaultSnippetSampling()),
	)
	if err == nil {
		t.Fatalf("expected invalid selector failure")
	}
}

func TestGenerateSnippetDatasetRejectsNonRealTarget(t *testing.T) {
	_, err := GenerateSnippetDataset(
		SnippetTargetSpec{
			TargetID:   "bad_target",
			TargetName: "Bad Target",
			SourceKind: SourceKindRaw,
			SourceRef:  "eml(1, x)",
			Selectors: []SnippetSelector{
				{Name: "x_leaf", PreorderIndex: 1},
			},
		},
		concepts.StandardLibrary(),
		[]SamplingDomain{
			{
				DomainID: "negative",
				Sampling: SamplingSpec{
					Variable:    "x",
					Start:       -1.0,
					Stop:        -0.1,
					PointCount:  4,
					SampleCount: 1,
					Seed:        0,
				},
			},
		},
	)
	if err == nil {
		t.Fatalf("expected non-real target failure")
	}
}

func TestWriteCuratedSnippetDatasets(t *testing.T) {
	root := t.TempDir()
	paths, artifacts, err := WriteCuratedSnippetDatasets(root)
	if err != nil {
		t.Fatalf("WriteCuratedSnippetDatasets returned error: %v", err)
	}
	if len(paths) != len(CuratedSnippetTargets()) || len(artifacts) != len(CuratedSnippetTargets()) {
		t.Fatalf("unexpected artifact counts: %d paths %d artifacts", len(paths), len(artifacts))
	}
	for _, path := range paths {
		if filepath.Dir(path) != filepath.Join(root, "artifacts", "snippets") {
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
