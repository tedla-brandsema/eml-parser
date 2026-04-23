package experiment

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSpecPathGeneratesDatasetAndResult(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "experiments", "specs"),
		filepath.Join(root, "experiments", "datasets"),
		filepath.Join(root, "experiments", "results"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
	}

	specPath := filepath.Join(root, "experiments", "specs", "exp_real_grid.json")
	specJSON := `{
	  "id": "exp_real_grid",
	  "description": "Recover exp(x) from a small evenly spaced real grid.",
	  "target": {
	    "kind": "concept",
	    "concept": "exp"
	  },
	  "dataset": {
	    "mode": "real_grid",
	    "variable": "x",
	    "grid": {
	      "start": -1.0,
	      "stop": 1.0,
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
	}`
	if err := os.WriteFile(specPath, []byte(specJSON), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	resultPath, artifact, err := RunSpecPath(root, specPath)
	if err != nil {
		t.Fatalf("RunSpecPath returned error: %v", err)
	}
	if artifact.ExperimentID != "exp_real_grid" {
		t.Fatalf("unexpected artifact: %+v", artifact)
	}
	if artifact.RecoveryStatus != recoveryPending {
		t.Fatalf("unexpected recovery status: %+v", artifact)
	}
	if len(artifact.Candidates) == 0 {
		t.Fatalf("expected candidates, got %+v", artifact)
	}
	if artifact.Candidates[0].CanonicalKey != "eml(x, 1)" {
		t.Fatalf("expected exp candidate first, got %+v", artifact.Candidates[0])
	}
	if !strings.HasSuffix(resultPath, filepath.Join("experiments", "results", "exp_real_grid.json")) {
		t.Fatalf("unexpected result path: %s", resultPath)
	}

	resultData, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("ReadFile result failed: %v", err)
	}
	if !strings.Contains(string(resultData), `"experiment_id": "exp_real_grid"`) {
		t.Fatalf("unexpected result contents: %q", string(resultData))
	}

	datasetPath := filepath.Join(root, "experiments", "datasets", "exp_real_grid.json")
	if _, err := os.Stat(datasetPath); err != nil {
		t.Fatalf("expected generated dataset at %s: %v", datasetPath, err)
	}
}

func TestRunSpecPathLoadsExistingDataset(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "experiments", "specs"),
		filepath.Join(root, "experiments", "datasets"),
		filepath.Join(root, "experiments", "results"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
	}

	specPath := filepath.Join(root, "experiments", "specs", "identity.json")
	specJSON := `{
	  "id": "identity",
	  "description": "Recover x from explicit points.",
	  "target": {
	    "kind": "raw",
	    "raw_eml": "x"
	  },
	  "dataset": {
	    "mode": "real_points",
	    "variable": "x",
	    "points": [-1, 0, 1]
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
	}`
	if err := os.WriteFile(specPath, []byte(specJSON), 0o644); err != nil {
		t.Fatalf("WriteFile spec failed: %v", err)
	}

	datasetPath := filepath.Join(root, "experiments", "datasets", "identity.json")
	datasetJSON := `{
	  "experiment_id": "identity",
	  "description": "Recover x from explicit points.",
	  "target": {
	    "kind": "raw",
	    "raw_eml": "x",
	    "canonical_key": "x",
	    "expression": "x"
	  },
	  "variable": "x",
	  "mode": "real_points",
	  "domain": {
	    "points": [-1, 0, 1]
	  },
	  "sample_count": 3,
	  "samples": [
	    {"input": -1, "target": -1},
	    {"input": 0, "target": 0},
	    {"input": 1, "target": 1}
	  ]
	}`
	if err := os.WriteFile(datasetPath, []byte(datasetJSON), 0o644); err != nil {
		t.Fatalf("WriteFile dataset failed: %v", err)
	}

	resultPath, artifact, err := RunSpecPath(root, specPath)
	if err != nil {
		t.Fatalf("RunSpecPath returned error: %v", err)
	}
	if artifact.DatasetPath != datasetPath {
		t.Fatalf("expected harness to use existing dataset path, got %+v", artifact)
	}
	if !strings.HasSuffix(resultPath, filepath.Join("experiments", "results", "identity.json")) {
		t.Fatalf("unexpected result path: %s", resultPath)
	}
}
