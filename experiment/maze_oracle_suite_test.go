package experiment

import (
	"path/filepath"
	"testing"
)

func TestMazeOracleSuiteSpecsMatchExpectedRecovery(t *testing.T) {
	projectRoot := ".."
	specPaths := []string{
		filepath.Join(projectRoot, "experiments", "specs", "maze_oracle_exp3_full_from_exp2.json"),
		filepath.Join(projectRoot, "experiments", "specs", "maze_oracle_exp3_snippet_from_exp1.json"),
		filepath.Join(projectRoot, "experiments", "specs", "maze_oracle_sinh_partial_coverage.json"),
		filepath.Join(projectRoot, "experiments", "specs", "maze_oracle_sigmoid_negative.json"),
	}

	for _, specPath := range specPaths {
		t.Run(filepath.Base(specPath), func(t *testing.T) {
			spec, err := LoadSpec(specPath)
			if err != nil {
				t.Fatalf("LoadSpec returned error: %v", err)
			}
			dataset, err := BuildDataset(spec)
			if err != nil {
				t.Fatalf("BuildDataset returned error: %v", err)
			}
			report, _, err := runMazeFromDataset(projectRoot, spec, dataset)
			if err != nil {
				t.Fatalf("runMazeFromDataset returned error: %v", err)
			}
			got := ClassifyMazeRecovery(spec, report)
			if got != spec.Recovery.ExpectedClass {
				t.Fatalf("expected recovery %q, got %q", spec.Recovery.ExpectedClass, got)
			}
		})
	}
}
