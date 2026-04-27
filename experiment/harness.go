package experiment

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"eml-parser/eval"
	"eml-parser/search"
)

// SearchResultArtifact is the machine-readable output of one oracle experiment
// run over the current enumerative real search path.
type SearchResultArtifact struct {
	ExperimentID       string              `json:"experiment_id"`
	Description        string              `json:"description"`
	SpecPath           string              `json:"spec_path,omitempty"`
	DatasetPath        string              `json:"dataset_path,omitempty"`
	Target             DatasetTarget       `json:"target"`
	TargetCanonicalKey string              `json:"target_canonical_key"`
	Dataset            DatasetMetadata     `json:"dataset"`
	Search             SearchExecution     `json:"search"`
	Diagnostics        DiagnosticsArtifact `json:"diagnostics"`
	Candidates         []CandidateResult   `json:"candidates"`
	RecoveryStatus     string              `json:"recovery_status"`
	CodeVersion        CodeVersion         `json:"code_version"`
	GeneratedAtUTC     string              `json:"generated_at_utc"`
}

// SearchExecution captures the concrete search configuration used for a run.
type SearchExecution struct {
	Mode   string     `json:"mode"`
	Bounds BoundsSpec `json:"bounds"`
	TopN   int        `json:"top_n"`
}

// DatasetMetadata is the dataset-side provenance retained in a search result.
type DatasetMetadata struct {
	Variable    string        `json:"variable"`
	Mode        string        `json:"mode"`
	Domain      DatasetDomain `json:"domain"`
	SampleCount int           `json:"sample_count"`
}

// CodeVersion identifies the code revision that produced a result artifact.
type CodeVersion struct {
	GitCommit string `json:"git_commit,omitempty"`
}

// DiagnosticsArtifact is the JSON-safe diagnostic view of one search run.
type DiagnosticsArtifact struct {
	GeneratedCount        int      `json:"generated_count"`
	UniqueCount           int      `json:"unique_count"`
	DuplicateCount        int      `json:"duplicate_count"`
	NormalizationHits     int      `json:"normalization_hits"`
	EvaluationRejects     int      `json:"evaluation_rejects"`
	NonFiniteCount        int      `json:"non_finite_count"`
	ScoredCount           int      `json:"scored_count"`
	ReturnedCount         int      `json:"returned_count"`
	BestScore             string   `json:"best_score"`
	WorstScore            string   `json:"worst_score"`
	MeanScore             string   `json:"mean_score"`
	TopCandidateSummaries []string `json:"top_candidate_summaries"`
}

// CandidateResult records one ranked candidate from a harness run.
type CandidateResult struct {
	Rank           int    `json:"rank"`
	Score          string `json:"score"`
	CanonicalKey   string `json:"canonical_key"`
	NormalizedExpr string `json:"normalized_expr"`
}

// LoadDataset reads a previously generated dataset artifact from disk.
func LoadDataset(path string) (DatasetArtifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DatasetArtifact{}, fmt.Errorf("read dataset artifact: %w", err)
	}
	var artifact DatasetArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return DatasetArtifact{}, fmt.Errorf("decode dataset artifact: %w", err)
	}
	return artifact, nil
}

// RunSpecPath executes one oracle experiment spec from disk, generating or
// loading its dataset and persisting a result artifact.
func RunSpecPath(projectRoot, specPath string) (string, SearchResultArtifact, error) {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return "", SearchResultArtifact{}, err
	}
	datasetPath := filepath.Join(projectRoot, "experiments", "datasets", spec.ID+".json")
	dataset, err := ensureDataset(projectRoot, datasetPath, spec)
	if err != nil {
		return "", SearchResultArtifact{}, err
	}

	report, err := runSearchFromDataset(spec, dataset)
	if err != nil {
		return "", SearchResultArtifact{}, err
	}

	artifact := SearchResultArtifact{
		ExperimentID:       spec.ID,
		Description:        spec.Description,
		SpecPath:           specPath,
		DatasetPath:        datasetPath,
		Target:             dataset.Target,
		TargetCanonicalKey: dataset.Target.CanonicalKey,
		Dataset: DatasetMetadata{
			Variable:    dataset.Variable,
			Mode:        dataset.Mode,
			Domain:      cloneDatasetDomain(dataset.Domain),
			SampleCount: dataset.SampleCount,
		},
		Search: SearchExecution{
			Mode:   spec.Search.Mode,
			Bounds: spec.Search.Bounds,
			TopN:   spec.Search.TopN,
		},
		Diagnostics:    diagnosticsArtifact(report.Diagnostics),
		Candidates:     candidateResults(report.Results),
		RecoveryStatus: ClassifyRecovery(spec, report),
		CodeVersion:    detectCodeVersion(projectRoot),
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
	}

	outputPath := filepath.Join(projectRoot, "experiments", "results", spec.ID+".json")
	if err := writeResultArtifact(outputPath, artifact); err != nil {
		return "", SearchResultArtifact{}, err
	}
	return outputPath, artifact, nil
}

func ensureDataset(projectRoot, datasetPath string, spec Spec) (DatasetArtifact, error) {
	if _, err := os.Stat(datasetPath); err == nil {
		return LoadDataset(datasetPath)
	} else if !os.IsNotExist(err) {
		return DatasetArtifact{}, fmt.Errorf("stat dataset artifact: %w", err)
	}
	_, artifact, err := WriteDataset(projectRoot, spec)
	if err != nil {
		return DatasetArtifact{}, err
	}
	return artifact, nil
}

func runSearchFromDataset(spec Spec, dataset DatasetArtifact) (search.SearchReport, error) {
	fixture := search.BenchmarkCase[float64]{
		Name:      spec.ID,
		Expr:      nil,
		Samples:   datasetSamplesToSearch(dataset),
		TargetKey: dataset.Variable,
	}
	return search.EnumerativeRealSearch(fixture, eval.Complex128Backend{}, search.SearchOptions{
		Bounds: search.Bounds{
			MaxDepth: spec.Search.Bounds.MaxDepth,
			MaxNodes: spec.Search.Bounds.MaxNodes,
		},
		TopN: spec.Search.TopN,
	})
}

func datasetSamplesToSearch(dataset DatasetArtifact) []search.Sample[float64] {
	out := make([]search.Sample[float64], 0, len(dataset.Samples))
	for _, sample := range dataset.Samples {
		out = append(out, search.Sample[float64]{
			Vars: map[string]float64{
				dataset.Variable: sample.Input,
			},
			Target: sample.Target,
		})
	}
	return out
}

func candidateResults(results []search.SearchResult) []CandidateResult {
	out := make([]CandidateResult, 0, len(results))
	for i, result := range results {
		out = append(out, CandidateResult{
			Rank:           i + 1,
			Score:          formatScore(result.Score),
			CanonicalKey:   result.Candidate.Key,
			NormalizedExpr: result.Candidate.Normalized.String(),
		})
	}
	return out
}

func cloneDatasetDomain(domain DatasetDomain) DatasetDomain {
	out := DatasetDomain{}
	if domain.Grid != nil {
		grid := *domain.Grid
		out.Grid = &grid
	}
	if len(domain.Points) > 0 {
		out.Points = append([]float64(nil), domain.Points...)
	}
	return out
}

func diagnosticsArtifact(d search.SearchDiagnostics) DiagnosticsArtifact {
	return DiagnosticsArtifact{
		GeneratedCount:        d.GeneratedCount,
		UniqueCount:           d.UniqueCount,
		DuplicateCount:        d.DuplicateCount,
		NormalizationHits:     d.NormalizationHits,
		EvaluationRejects:     d.EvaluationRejects,
		NonFiniteCount:        d.NonFiniteCount,
		ScoredCount:           d.ScoredCount,
		ReturnedCount:         d.ReturnedCount,
		BestScore:             formatScore(d.BestScore),
		WorstScore:            formatScore(d.WorstScore),
		MeanScore:             formatScore(d.MeanScore),
		TopCandidateSummaries: append([]string(nil), d.TopCandidateSummaries...),
	}
}

func (d DiagnosticsArtifact) String() string {
	return fmt.Sprintf(
		"generated: %d\nunique: %d\nduplicates: %d\nnormalization_hits: %d\nevaluation_rejects: %d\nnon_finite_count: %d\nscored: %d\nreturned: %d\nbest_score: %s\nworst_score: %s\nmean_score: %s",
		d.GeneratedCount,
		d.UniqueCount,
		d.DuplicateCount,
		d.NormalizationHits,
		d.EvaluationRejects,
		d.NonFiniteCount,
		d.ScoredCount,
		d.ReturnedCount,
		d.BestScore,
		d.WorstScore,
		d.MeanScore,
	)
}

func formatScore(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "+Inf"
	case math.IsInf(v, -1):
		return "-Inf"
	default:
		return fmt.Sprintf("%g", v)
	}
}

func detectCodeVersion(projectRoot string) CodeVersion {
	cmd := exec.Command("git", "-C", projectRoot, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return CodeVersion{}
	}
	commit := strings.TrimSpace(string(output))
	if commit == "" {
		return CodeVersion{}
	}
	return CodeVersion{GitCommit: commit}
}

func writeResultArtifact(path string, artifact SearchResultArtifact) error {
	payload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result artifact: %w", err)
	}
	if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
		return fmt.Errorf("write result artifact: %w", err)
	}
	return nil
}
