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
	"eml-parser/family"
	"eml-parser/search"
	"eml-parser/search/maze"
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

	// Maze-mode extensions. All empty for enumerative runs.
	Anchors         []AnchorArtifact         `json:"anchors,omitempty"`
	MazeDiagnostics *MazeDiagnosticsArtifact `json:"maze_diagnostics,omitempty"`
	PartialResults  []PartialResultArtifact  `json:"partial_results,omitempty"`
}

// SearchExecution captures the concrete search configuration used for a run.
type SearchExecution struct {
	Mode   string     `json:"mode"`
	Bounds BoundsSpec `json:"bounds"`
	TopN   int        `json:"top_n"`
	Maze   *MazeSpec  `json:"maze,omitempty"`
}

// AnchorArtifact records one declared maze anchor and its snippet provenance.
type AnchorArtifact struct {
	Name       string                 `json:"name"`
	Expression string                 `json:"expression"`
	Provenance *maze.AnchorProvenance `json:"provenance,omitempty"`
}

// MazeDiagnosticsArtifact is the JSON-safe diagnostic view of one maze run.
type MazeDiagnosticsArtifact struct {
	AnchorCount                int    `json:"anchor_count"`
	ThreadsSpawned             int    `json:"threads_spawned"`
	BranchesExpanded           int    `json:"branches_expanded"`
	BranchesPruned             int    `json:"branches_pruned"`
	BranchesRetained           int    `json:"branches_retained"`
	BranchesCompleted          int    `json:"branches_completed"`
	DuplicateEliminations      int    `json:"duplicate_eliminations"`
	FrontierExpansionsTried    int    `json:"frontier_expansions_tried"`
	FrontierExpansionsAccepted int    `json:"frontier_expansions_accepted"`
	FrontierExpansionsRejected int    `json:"frontier_expansions_rejected"`
	MaxDepthReached            int    `json:"max_depth_reached"`
	MaxFrontierCountSeen       int    `json:"max_frontier_count_seen"`
	BestScore                  string `json:"best_score"`
}

// PartialResultArtifact records one retained partial maze result.
type PartialResultArtifact struct {
	AnchorName     string  `json:"anchor_name"`
	CanonicalKey   string  `json:"canonical_key"`
	NormalizedExpr string  `json:"normalized_expr"`
	Score          string  `json:"score"`
	Reason         string  `json:"reason"`
	CoverageRatio  float64 `json:"coverage_ratio,omitempty"`
	CoveredCount   int     `json:"covered_count,omitempty"`
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

// CandidateResult records one ranked candidate from a harness run. The
// coverage and anchor fields are populated only by maze-mode runs; Windows is
// populated only under window-set coverage scoring.
type CandidateResult struct {
	Rank           int              `json:"rank"`
	Score          string           `json:"score"`
	CanonicalKey   string           `json:"canonical_key"`
	NormalizedExpr string           `json:"normalized_expr"`
	AnchorName     string           `json:"anchor_name,omitempty"`
	CoverageRatio  float64          `json:"coverage_ratio,omitempty"`
	CoveredCount   int              `json:"covered_count,omitempty"`
	WindowStart    int              `json:"window_start,omitempty"`
	WindowEnd      int              `json:"window_end,omitempty"`
	LocalError     string           `json:"local_error,omitempty"`
	Windows        []WindowArtifact `json:"windows,omitempty"`
}

// WindowArtifact records one explained window of a window-set scored
// candidate over the sorted sample trace.
type WindowArtifact struct {
	Start        int    `json:"start"`
	End          int    `json:"end"`
	CoveredCount int    `json:"covered_count"`
	LocalError   string `json:"local_error"`
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
			Maze:   spec.Search.Maze,
		},
		CodeVersion:    detectCodeVersion(projectRoot),
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
	}

	switch spec.Search.Mode {
	case SearchModeMazeReal:
		report, anchors, err := runMazeFromDataset(projectRoot, spec, dataset)
		if err != nil {
			return "", SearchResultArtifact{}, err
		}
		artifact.Anchors = anchorArtifacts(anchors)
		artifact.MazeDiagnostics = mazeDiagnosticsArtifact(report.Diagnostics)
		artifact.PartialResults = partialResultArtifacts(report.PartialResults)
		artifact.Candidates = mazeCandidateResults(report.BestCandidates)
		artifact.RecoveryStatus = ClassifyMazeRecovery(spec, report)
	default:
		report, err := runSearchFromDataset(spec, dataset)
		if err != nil {
			return "", SearchResultArtifact{}, err
		}
		artifact.Diagnostics = diagnosticsArtifact(report.Diagnostics)
		artifact.Candidates = candidateResults(report.Results)
		artifact.RecoveryStatus = ClassifyRecovery(spec, report)
	}

	outputPath := filepath.Join(projectRoot, "experiments", "results", spec.ID+".json")
	if err := writeResultArtifact(outputPath, artifact); err != nil {
		return "", SearchResultArtifact{}, err
	}
	return outputPath, artifact, nil
}

// runMazeFromDataset executes one maze_real experiment over a generated
// dataset, seeding anchors from the spec's committed snippet artifact.
func runMazeFromDataset(projectRoot string, spec Spec, dataset DatasetArtifact) (maze.MazeReport, []maze.Anchor, error) {
	snippetArtifact, err := ensureSnippetArtifact(projectRoot, spec.Search.Maze.SnippetArtifact)
	if err != nil {
		return maze.MazeReport{}, nil, err
	}
	anchors, err := maze.AnchorsFromSnippetArtifact(snippetArtifact, spec.Search.Maze.SnippetIDs...)
	if err != nil {
		return maze.MazeReport{}, nil, err
	}

	fixture := search.BenchmarkCase[float64]{
		Name:      spec.ID,
		Expr:      nil,
		Samples:   datasetSamplesToSearch(dataset),
		TargetKey: dataset.Variable,
	}
	options := maze.MazeOptions{
		Bounds: search.Bounds{
			MaxDepth: spec.Search.Bounds.MaxDepth,
			MaxNodes: spec.Search.Bounds.MaxNodes,
		},
		TopN:            spec.Search.TopN,
		AcceptThreshold: spec.Search.Maze.AcceptThreshold,
		RetainThreshold: spec.Search.Maze.RetainThreshold,
		MinImprovement:  spec.Search.Maze.MinImprovement,
	}

	var report maze.MazeReport
	switch coverage := spec.Search.Maze.Coverage; {
	case coverage == nil:
		report, err = maze.MazeRealSearch(fixture, eval.Complex128Backend{}, anchors, options)
	case coverage.Mode == CoverageModeWindowSet:
		report, err = maze.MazeRealSearchWindowSet(fixture, eval.Complex128Backend{}, anchors, options, maze.WindowSetOptions{
			PointTolerance: coverage.PointTolerance,
			MinWindowSize:  coverage.MinWindowSize,
			MaxWindowCount: coverage.MaxWindowCount,
			CoverageWeight: coverage.CoverageWeight,
		})
	default:
		report, err = maze.MazeRealSearchPartialCoverage(fixture, eval.Complex128Backend{}, anchors, options, maze.CoverageOptions{
			MinWindowSize:  coverage.MinWindowSize,
			MaxWindowSize:  coverage.MaxWindowSize,
			CoverageWeight: coverage.CoverageWeight,
		})
	}
	if err != nil {
		return maze.MazeReport{}, nil, err
	}
	return report, anchors, nil
}

// ensureSnippetArtifact loads the referenced snippet artifact, regenerating
// the curated snippet corpus once and retrying if the artifact is missing or
// unreadable. Curated generation is deterministic, so regeneration is always
// safe.
func ensureSnippetArtifact(projectRoot, relativePath string) (family.SnippetDatasetArtifact, error) {
	path := filepath.Join(projectRoot, relativePath)
	artifact, err := maze.LoadSnippetArtifact(path)
	if err == nil {
		return artifact, nil
	}
	if _, _, genErr := family.WriteCuratedSnippetDatasets(projectRoot); genErr != nil {
		return family.SnippetDatasetArtifact{}, fmt.Errorf("regenerate curated snippet datasets: %w", genErr)
	}
	return maze.LoadSnippetArtifact(path)
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

func anchorArtifacts(anchors []maze.Anchor) []AnchorArtifact {
	out := make([]AnchorArtifact, 0, len(anchors))
	for _, anchor := range anchors {
		out = append(out, AnchorArtifact{
			Name:       anchor.Name,
			Expression: anchor.Expr.String(),
			Provenance: anchor.Provenance,
		})
	}
	return out
}

func mazeDiagnosticsArtifact(d maze.MazeDiagnostics) *MazeDiagnosticsArtifact {
	return &MazeDiagnosticsArtifact{
		AnchorCount:                d.AnchorCount,
		ThreadsSpawned:             d.ThreadsSpawned,
		BranchesExpanded:           d.BranchesExpanded,
		BranchesPruned:             d.BranchesPruned,
		BranchesRetained:           d.BranchesRetained,
		BranchesCompleted:          d.BranchesCompleted,
		DuplicateEliminations:      d.DuplicateEliminations,
		FrontierExpansionsTried:    d.FrontierExpansionsTried,
		FrontierExpansionsAccepted: d.FrontierExpansionsAccepted,
		FrontierExpansionsRejected: d.FrontierExpansionsRejected,
		MaxDepthReached:            d.MaxDepthReached,
		MaxFrontierCountSeen:       d.MaxFrontierCountSeen,
		BestScore:                  formatScore(d.BestScore),
	}
}

func mazeCandidateResults(candidates []maze.CandidateScore) []CandidateResult {
	out := make([]CandidateResult, 0, len(candidates))
	for i, candidate := range candidates {
		out = append(out, CandidateResult{
			Rank:           i + 1,
			Score:          formatScore(candidate.Score),
			CanonicalKey:   candidate.Candidate.Key,
			NormalizedExpr: candidate.Candidate.Normalized.String(),
			AnchorName:     candidate.AnchorName,
			CoverageRatio:  candidate.ScoreDetails.CoverageRatio,
			CoveredCount:   candidate.ScoreDetails.CoveredCount,
			WindowStart:    candidate.ScoreDetails.WindowStart,
			WindowEnd:      candidate.ScoreDetails.WindowEnd,
			LocalError:     formatScore(candidate.ScoreDetails.LocalError),
			Windows:        windowArtifacts(candidate.ScoreDetails.Windows),
		})
	}
	return out
}

func windowArtifacts(windows []search.CoverageWindow) []WindowArtifact {
	if len(windows) == 0 {
		return nil
	}
	out := make([]WindowArtifact, 0, len(windows))
	for _, window := range windows {
		out = append(out, WindowArtifact{
			Start:        window.Start,
			End:          window.End,
			CoveredCount: window.CoveredCount,
			LocalError:   formatScore(window.LocalError),
		})
	}
	return out
}

func partialResultArtifacts(partials []maze.PartialResult) []PartialResultArtifact {
	out := make([]PartialResultArtifact, 0, len(partials))
	for _, partial := range partials {
		out = append(out, PartialResultArtifact{
			AnchorName:     partial.AnchorName,
			CanonicalKey:   partial.Candidate.Key,
			NormalizedExpr: partial.Candidate.Normalized.String(),
			Score:          formatScore(partial.Score),
			Reason:         partial.Reason,
			CoverageRatio:  partial.ScoreDetails.CoverageRatio,
			CoveredCount:   partial.ScoreDetails.CoveredCount,
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

func (d MazeDiagnosticsArtifact) String() string {
	return fmt.Sprintf(
		"anchors: %d\nthreads_spawned: %d\nbranches_expanded: %d\nbranches_pruned: %d\nbranches_retained: %d\nbranches_completed: %d\nduplicate_eliminations: %d\nfrontier_expansions_tried: %d\nfrontier_expansions_accepted: %d\nfrontier_expansions_rejected: %d\nmax_depth_reached: %d\nmax_frontier_count_seen: %d\nbest_score: %s",
		d.AnchorCount,
		d.ThreadsSpawned,
		d.BranchesExpanded,
		d.BranchesPruned,
		d.BranchesRetained,
		d.BranchesCompleted,
		d.DuplicateEliminations,
		d.FrontierExpansionsTried,
		d.FrontierExpansionsAccepted,
		d.FrontierExpansionsRejected,
		d.MaxDepthReached,
		d.MaxFrontierCountSeen,
		d.BestScore,
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
