package experiment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildSuiteSummary(t *testing.T) {
	root := t.TempDir()
	resultA := filepath.Join(root, "a.json")
	resultB := filepath.Join(root, "b.json")

	writeTestResult(t, resultA, SearchResultArtifact{
		ExperimentID:       "exp_a",
		Description:        "positive control",
		Target:             DatasetTarget{Kind: TargetKindConcept, Concept: "exp"},
		TargetCanonicalKey: "eml(x, 1)",
		Dataset:            DatasetMetadata{Variable: "x", Mode: DatasetModeRealGrid, SampleCount: 5},
		Search:             SearchExecution{Mode: SearchModeEnumerativeReal, Bounds: BoundsSpec{MaxDepth: 2, MaxNodes: 3}, TopN: 5},
		Diagnostics: DiagnosticsArtifact{
			GeneratedCount:    10,
			UniqueCount:       8,
			ReturnedCount:     5,
			EvaluationRejects: 1,
		},
		Candidates: []CandidateResult{{
			Rank:           1,
			Score:          "0",
			CanonicalKey:   "eml(x, 1)",
			NormalizedExpr: "eml(x, 1)",
		}},
		RecoveryStatus: RecoveryClassExactNormalized,
	})
	writeTestResult(t, resultB, SearchResultArtifact{
		ExperimentID:       "raw_b",
		Description:        "negative control",
		Target:             DatasetTarget{Kind: TargetKindRaw, RawEML: "x"},
		TargetCanonicalKey: "x",
		Dataset:            DatasetMetadata{Variable: "x", Mode: DatasetModeExplicitPoints, SampleCount: 3},
		Search:             SearchExecution{Mode: SearchModeEnumerativeReal, Bounds: BoundsSpec{MaxDepth: 1, MaxNodes: 1}, TopN: 3},
		Diagnostics: DiagnosticsArtifact{
			GeneratedCount:    3,
			UniqueCount:       3,
			ReturnedCount:     1,
			EvaluationRejects: 0,
		},
		Candidates: []CandidateResult{{
			Rank:           1,
			Score:          "1",
			CanonicalKey:   "1",
			NormalizedExpr: "1",
		}},
		RecoveryStatus: RecoveryClassNoRecovery,
	})

	summary, err := BuildSuiteSummary("smoke_suite", []string{resultB, resultA})
	if err != nil {
		t.Fatalf("BuildSuiteSummary returned error: %v", err)
	}
	if summary.TotalExperiments != 2 || summary.SuccessCount != 1 || summary.FailureCount != 1 {
		t.Fatalf("unexpected counts: %+v", summary)
	}
	if summary.RecoveryClassCounts[RecoveryClassExactNormalized] != 1 || summary.RecoveryClassCounts[RecoveryClassNoRecovery] != 1 {
		t.Fatalf("unexpected recovery class counts: %+v", summary.RecoveryClassCounts)
	}
	if summary.TargetFamilyCounts["exp"] != 1 || summary.TargetFamilyCounts[TargetKindRaw] != 1 {
		t.Fatalf("unexpected target family counts: %+v", summary.TargetFamilyCounts)
	}
	if summary.Diagnostics.GeneratedCount.Min != 3 || summary.Diagnostics.GeneratedCount.Max != 10 {
		t.Fatalf("unexpected generated range: %+v", summary.Diagnostics)
	}
	if len(summary.Examples) != 2 || summary.Examples[0].ExperimentID != "exp_a" {
		t.Fatalf("unexpected examples: %+v", summary.Examples)
	}
}

func TestWriteSuiteReports(t *testing.T) {
	root := t.TempDir()
	reportsDir := filepath.Join(root, "experiments", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	resultPath := filepath.Join(root, "result.json")
	writeTestResult(t, resultPath, SearchResultArtifact{
		ExperimentID:   "exp_only",
		Target:         DatasetTarget{Kind: TargetKindConcept, Concept: "exp"},
		RecoveryStatus: RecoveryClassExactNormalized,
		Diagnostics: DiagnosticsArtifact{
			GeneratedCount:    4,
			UniqueCount:       4,
			ReturnedCount:     2,
			EvaluationRejects: 0,
		},
		Candidates: []CandidateResult{{
			Rank:           1,
			Score:          "0",
			CanonicalKey:   "eml(x, 1)",
			NormalizedExpr: "eml(x, 1)",
		}},
	})

	jsonPath, mdPath, summary, err := WriteSuiteReports(root, "oracle_smoke", []string{resultPath})
	if err != nil {
		t.Fatalf("WriteSuiteReports returned error: %v", err)
	}
	if summary.TotalExperiments != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("ReadFile json failed: %v", err)
	}
	if !strings.Contains(string(jsonData), `"suite_id": "oracle_smoke"`) {
		t.Fatalf("unexpected json report: %q", string(jsonData))
	}
	mdData, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("ReadFile markdown failed: %v", err)
	}
	for _, expected := range []string{
		"# Suite oracle_smoke",
		"## Recovery Classes",
		"## Top Recovered Expressions",
	} {
		if !strings.Contains(string(mdData), expected) {
			t.Fatalf("expected %q in markdown report, got %q", expected, string(mdData))
		}
	}
}

func writeTestResult(t *testing.T, path string, artifact SearchResultArtifact) {
	t.Helper()
	payload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
