package experiment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDatasetForConceptGrid(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "exp_real_grid",
		"description": "Recover exp(x) on a small real grid",
		"target": {
			"kind": "concept",
			"concept": "exp"
		},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {
				"start": -1,
				"stop": 1,
				"count": 5
			}
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 2,
				"max_nodes": 3
			},
			"top_n": 5
		},
		"recovery": {
			"expected_class": "exact_normalized_recovery",
			"expected_canonical_key": "eml(x, 1)"
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}

	artifact, err := BuildDataset(spec)
	if err != nil {
		t.Fatalf("BuildDataset returned error: %v", err)
	}
	if artifact.ExperimentID != spec.ID {
		t.Fatalf("unexpected dataset artifact id: %+v", artifact)
	}
	if artifact.Target.Kind != TargetKindConcept || artifact.Target.Concept != "exp" {
		t.Fatalf("unexpected target metadata: %+v", artifact.Target)
	}
	if artifact.Target.CanonicalKey != "eml(x, 1)" {
		t.Fatalf("unexpected canonical key: %+v", artifact.Target)
	}
	if artifact.SampleCount != 5 || len(artifact.Samples) != 5 {
		t.Fatalf("unexpected sample count: %+v", artifact)
	}
	if artifact.Domain.Grid == nil || artifact.Domain.Grid.Count != 5 {
		t.Fatalf("unexpected grid metadata: %+v", artifact.Domain)
	}
}

func TestBuildDatasetForRawPoints(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "raw_identity_points",
		"description": "Evaluate x over explicit points",
		"target": {
			"kind": "raw",
			"raw_eml": "x"
		},
		"dataset": {
			"mode": "real_points",
			"variable": "x",
			"points": [-2, 0, 3]
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 1,
				"max_nodes": 1
			},
			"top_n": 3
		},
		"recovery": {
			"expected_class": "exact_normalized_recovery",
			"expected_canonical_key": "x"
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}

	artifact, err := BuildDataset(spec)
	if err != nil {
		t.Fatalf("BuildDataset returned error: %v", err)
	}
	if artifact.Target.Kind != TargetKindRaw || artifact.Target.RawEML != "x" {
		t.Fatalf("unexpected raw target metadata: %+v", artifact.Target)
	}
	if artifact.Target.Expression != "x" {
		t.Fatalf("unexpected normalized target expression: %+v", artifact.Target)
	}
	if len(artifact.Domain.Points) != 3 {
		t.Fatalf("unexpected points metadata: %+v", artifact.Domain)
	}
	for i, sample := range artifact.Samples {
		if sample.Target != artifact.Domain.Points[i] {
			t.Fatalf("expected identity dataset, got %+v", artifact.Samples)
		}
	}
}

func TestBuildDatasetRejectsNonRealOutputs(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "non_real_target",
		"description": "i is not real on the real line",
		"target": {
			"kind": "concept",
			"concept": "i"
		},
		"dataset": {
			"mode": "real_points",
			"variable": "x",
			"points": [0]
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 1,
				"max_nodes": 1
			},
			"top_n": 1
		},
		"recovery": {
			"expected_class": "no_recovery"
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}

	_, err = BuildDataset(spec)
	if err == nil {
		t.Fatal("expected non-real output error")
	}
	if !strings.Contains(err.Error(), "non-real output") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWriteDataset(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "write_dataset",
		"description": "Write a small exp dataset",
		"target": {
			"kind": "concept",
			"concept": "exp"
		},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {
				"start": 0,
				"stop": 1,
				"count": 2
			}
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 2,
				"max_nodes": 3
			},
			"top_n": 2
		},
		"recovery": {
			"expected_class": "exact_normalized_recovery",
			"expected_canonical_key": "eml(x, 1)"
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "experiments", "datasets"), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	outputPath, artifact, err := WriteDataset(root, spec)
	if err != nil {
		t.Fatalf("WriteDataset returned error: %v", err)
	}
	if !strings.HasSuffix(outputPath, filepath.Join("experiments", "datasets", "write_dataset.json")) {
		t.Fatalf("unexpected output path: %s", outputPath)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(data), `"experiment_id": "write_dataset"`) {
		t.Fatalf("unexpected dataset file contents: %q", string(data))
	}
	if artifact.SampleCount != 2 {
		t.Fatalf("unexpected dataset artifact: %+v", artifact)
	}
}
