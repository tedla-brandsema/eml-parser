package experiment

import (
	"path/filepath"
	"testing"
)

func TestInitialOracleSuiteSpecsMatchExpectedRecovery(t *testing.T) {
	projectRoot := "/home/ted/projects/go/eml-parser"
	specPaths := []string{
		filepath.Join(projectRoot, "experiments", "specs", "oracle_exp_exact.json"),
		filepath.Join(projectRoot, "experiments", "specs", "oracle_log_exact.json"),
		filepath.Join(projectRoot, "experiments", "specs", "oracle_exp_exp_exact.json"),
		filepath.Join(projectRoot, "experiments", "specs", "oracle_add_exp_x_negative.json"),
		filepath.Join(projectRoot, "experiments", "specs", "oracle_sin_negative.json"),
		filepath.Join(projectRoot, "experiments", "specs", "oracle_sigmoid_negative.json"),
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
			report, err := runSearchFromDataset(spec, dataset)
			if err != nil {
				t.Fatalf("runSearchFromDataset returned error: %v", err)
			}
			got := ClassifyRecovery(spec, report)
			if got != spec.Recovery.ExpectedClass {
				t.Fatalf("expected recovery %q, got %q", spec.Recovery.ExpectedClass, got)
			}
		})
	}
}
