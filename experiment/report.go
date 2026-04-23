package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SuiteSummary is the machine-readable aggregate view of a set of experiment
// result artifacts.
type SuiteSummary struct {
	SuiteID             string           `json:"suite_id"`
	ResultPaths         []string         `json:"result_paths"`
	TotalExperiments    int              `json:"total_experiments"`
	SuccessCount        int              `json:"success_count"`
	FailureCount        int              `json:"failure_count"`
	RecoveryClassCounts map[string]int   `json:"recovery_class_counts"`
	TargetFamilyCounts  map[string]int   `json:"target_family_counts"`
	Diagnostics         DiagnosticsRange `json:"diagnostics"`
	Examples            []SuiteExample   `json:"examples"`
	GeneratedAtUTC      string           `json:"generated_at_utc"`
}

// DiagnosticsRange summarizes aggregate diagnostic ranges over a suite.
type DiagnosticsRange struct {
	GeneratedCount Range `json:"generated_count"`
	UniqueCount    Range `json:"unique_count"`
	ReturnedCount  Range `json:"returned_count"`
	EvalRejects    Range `json:"evaluation_rejects"`
}

// Range is a simple min/max summary for integer diagnostics.
type Range struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// SuiteExample is the top recovered expression for one experiment in the suite.
type SuiteExample struct {
	ExperimentID    string `json:"experiment_id"`
	RecoveryStatus  string `json:"recovery_status"`
	TargetFamily    string `json:"target_family"`
	TopExpression   string `json:"top_expression,omitempty"`
	TopCanonicalKey string `json:"top_canonical_key,omitempty"`
	TopScore        string `json:"top_score,omitempty"`
}

// LoadResultArtifact reads a previously generated search result artifact.
func LoadResultArtifact(path string) (SearchResultArtifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SearchResultArtifact{}, fmt.Errorf("read result artifact: %w", err)
	}
	var artifact SearchResultArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return SearchResultArtifact{}, fmt.Errorf("decode result artifact: %w", err)
	}
	return artifact, nil
}

// BuildSuiteSummary aggregates a set of result artifact paths into one suite
// summary.
func BuildSuiteSummary(suiteID string, resultPaths []string) (SuiteSummary, error) {
	if strings.TrimSpace(suiteID) == "" {
		return SuiteSummary{}, fmt.Errorf("suite id cannot be empty")
	}
	if len(resultPaths) == 0 {
		return SuiteSummary{}, fmt.Errorf("suite %q requires at least one result path", suiteID)
	}

	sortedPaths := append([]string(nil), resultPaths...)
	sort.Strings(sortedPaths)

	summary := SuiteSummary{
		SuiteID:             suiteID,
		ResultPaths:         sortedPaths,
		RecoveryClassCounts: map[string]int{},
		TargetFamilyCounts:  map[string]int{},
		GeneratedAtUTC:      time.Now().UTC().Format(time.RFC3339),
	}

	var initialized bool
	for _, path := range sortedPaths {
		artifact, err := LoadResultArtifact(path)
		if err != nil {
			return SuiteSummary{}, err
		}
		summary.TotalExperiments++
		summary.RecoveryClassCounts[artifact.RecoveryStatus]++
		if artifact.RecoveryStatus == RecoveryClassNoRecovery {
			summary.FailureCount++
		} else {
			summary.SuccessCount++
		}

		family := targetFamily(artifact)
		summary.TargetFamilyCounts[family]++
		summary.Examples = append(summary.Examples, suiteExample(artifact, family))

		if !initialized {
			summary.Diagnostics = DiagnosticsRange{
				GeneratedCount: Range{Min: artifact.Diagnostics.GeneratedCount, Max: artifact.Diagnostics.GeneratedCount},
				UniqueCount:    Range{Min: artifact.Diagnostics.UniqueCount, Max: artifact.Diagnostics.UniqueCount},
				ReturnedCount:  Range{Min: artifact.Diagnostics.ReturnedCount, Max: artifact.Diagnostics.ReturnedCount},
				EvalRejects:    Range{Min: artifact.Diagnostics.EvaluationRejects, Max: artifact.Diagnostics.EvaluationRejects},
			}
			initialized = true
			continue
		}
		updateRange(&summary.Diagnostics.GeneratedCount, artifact.Diagnostics.GeneratedCount)
		updateRange(&summary.Diagnostics.UniqueCount, artifact.Diagnostics.UniqueCount)
		updateRange(&summary.Diagnostics.ReturnedCount, artifact.Diagnostics.ReturnedCount)
		updateRange(&summary.Diagnostics.EvalRejects, artifact.Diagnostics.EvaluationRejects)
	}

	sort.Slice(summary.Examples, func(i, j int) bool {
		return summary.Examples[i].ExperimentID < summary.Examples[j].ExperimentID
	})
	return summary, nil
}

// WriteSuiteReports writes the machine-readable and Markdown summaries for a
// named suite under experiments/reports.
func WriteSuiteReports(projectRoot, suiteID string, resultPaths []string) (string, string, SuiteSummary, error) {
	summary, err := BuildSuiteSummary(suiteID, resultPaths)
	if err != nil {
		return "", "", SuiteSummary{}, err
	}

	jsonPath := filepath.Join(projectRoot, "experiments", "reports", suiteID+".json")
	mdPath := filepath.Join(projectRoot, "experiments", "reports", suiteID+".md")

	jsonPayload, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", "", SuiteSummary{}, fmt.Errorf("marshal suite summary: %w", err)
	}
	if err := os.WriteFile(jsonPath, append(jsonPayload, '\n'), 0o644); err != nil {
		return "", "", SuiteSummary{}, fmt.Errorf("write suite summary json: %w", err)
	}

	if err := os.WriteFile(mdPath, []byte(summary.Markdown()), 0o644); err != nil {
		return "", "", SuiteSummary{}, fmt.Errorf("write suite summary markdown: %w", err)
	}

	return jsonPath, mdPath, summary, nil
}

// Markdown renders a lightweight human-readable summary for a suite.
func (s SuiteSummary) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Suite %s\n\n", s.SuiteID)
	fmt.Fprintf(&b, "- total_experiments: %d\n", s.TotalExperiments)
	fmt.Fprintf(&b, "- success_count: %d\n", s.SuccessCount)
	fmt.Fprintf(&b, "- failure_count: %d\n", s.FailureCount)
	fmt.Fprintf(&b, "- generated_at_utc: %s\n\n", s.GeneratedAtUTC)

	fmt.Fprintf(&b, "## Recovery Classes\n\n")
	for _, key := range sortedMapKeys(s.RecoveryClassCounts) {
		fmt.Fprintf(&b, "- %s: %d\n", key, s.RecoveryClassCounts[key])
	}
	fmt.Fprintf(&b, "\n## Target Families\n\n")
	for _, key := range sortedMapKeys(s.TargetFamilyCounts) {
		fmt.Fprintf(&b, "- %s: %d\n", key, s.TargetFamilyCounts[key])
	}
	fmt.Fprintf(&b, "\n## Diagnostic Ranges\n\n")
	fmt.Fprintf(&b, "- generated_count: %d..%d\n", s.Diagnostics.GeneratedCount.Min, s.Diagnostics.GeneratedCount.Max)
	fmt.Fprintf(&b, "- unique_count: %d..%d\n", s.Diagnostics.UniqueCount.Min, s.Diagnostics.UniqueCount.Max)
	fmt.Fprintf(&b, "- returned_count: %d..%d\n", s.Diagnostics.ReturnedCount.Min, s.Diagnostics.ReturnedCount.Max)
	fmt.Fprintf(&b, "- evaluation_rejects: %d..%d\n", s.Diagnostics.EvalRejects.Min, s.Diagnostics.EvalRejects.Max)

	fmt.Fprintf(&b, "\n## Top Recovered Expressions\n\n")
	for _, example := range s.Examples {
		fmt.Fprintf(
			&b,
			"- %s: class=%s family=%s expr=%s key=%s score=%s\n",
			example.ExperimentID,
			example.RecoveryStatus,
			example.TargetFamily,
			orNone(example.TopExpression),
			orNone(example.TopCanonicalKey),
			orNone(example.TopScore),
		)
	}
	return b.String()
}

func suiteExample(artifact SearchResultArtifact, family string) SuiteExample {
	example := SuiteExample{
		ExperimentID:   artifact.ExperimentID,
		RecoveryStatus: artifact.RecoveryStatus,
		TargetFamily:   family,
	}
	if len(artifact.Candidates) > 0 {
		example.TopExpression = artifact.Candidates[0].NormalizedExpr
		example.TopCanonicalKey = artifact.Candidates[0].CanonicalKey
		example.TopScore = artifact.Candidates[0].Score
	}
	return example
}

func targetFamily(artifact SearchResultArtifact) string {
	if artifact.Target.Concept != "" {
		return artifact.Target.Concept
	}
	return artifact.Target.Kind
}

func updateRange(r *Range, value int) {
	if value < r.Min {
		r.Min = value
	}
	if value > r.Max {
		r.Max = value
	}
}

func sortedMapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func orNone(v string) string {
	if v == "" {
		return "(none)"
	}
	return v
}
