package maze

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

	"eml-parser/eval"
	"eml-parser/family"
	"eml-parser/parser"
	"eml-parser/search/common"
)

type AnchorSourceKind string

const (
	AnchorSourceManual  AnchorSourceKind = "manual"
	AnchorSourceSnippet AnchorSourceKind = "snippet"
)

type AnchorProvenance struct {
	SourceKind     AnchorSourceKind `json:"source_kind"`
	TargetID       string           `json:"target_id,omitempty"`
	TargetName     string           `json:"target_name,omitempty"`
	ParentTargetID string           `json:"parent_target_id,omitempty"`
	SnippetID      string           `json:"snippet_id,omitempty"`
	PreorderIndex  int              `json:"preorder_index,omitempty"`
	TreePath       string           `json:"tree_path,omitempty"`
	CanonicalKey   string           `json:"canonical_key,omitempty"`
}

type SnippetMatch struct {
	TargetID       string           `json:"target_id"`
	TargetName     string           `json:"target_name"`
	ParentTargetID string           `json:"parent_target_id"`
	SnippetID      string           `json:"snippet_id"`
	CanonicalKey   string           `json:"canonical_key"`
	DomainID       string           `json:"domain_id"`
	SampleSetID    string           `json:"sample_set_id"`
	Score          float64          `json:"score"`
	Provenance     AnchorProvenance `json:"provenance"`
}

type SpawnOptions struct {
	TopK     int
	MaxScore float64
}

type SpawnDiagnostics struct {
	ArtifactsExamined     int
	SnippetTracesCompared int
	AnchorsPromoted       int
	ThresholdRejects      int
	ShapeRejects          int
	BestScore             float64
}

type SpawnReport struct {
	Matches     []SnippetMatch
	Anchors     []Anchor
	Diagnostics SpawnDiagnostics
}

// LoadSnippetArtifact reads one snippet artifact from disk.
func LoadSnippetArtifact(path string) (family.SnippetDatasetArtifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return family.SnippetDatasetArtifact{}, fmt.Errorf("read snippet artifact: %w", err)
	}
	var artifact family.SnippetDatasetArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return family.SnippetDatasetArtifact{}, fmt.Errorf("decode snippet artifact: %w", err)
	}
	return artifact, nil
}

// AnchorsFromSnippetArtifact converts selected snippet descriptors into explicit
// maze anchors while preserving snippet provenance.
func AnchorsFromSnippetArtifact(artifact family.SnippetDatasetArtifact, snippetIDs ...string) ([]Anchor, error) {
	if len(artifact.Snippets) == 0 {
		return nil, fmt.Errorf("snippet artifact %q has no snippets", artifact.TargetID)
	}
	selected := make(map[string]struct{}, len(snippetIDs))
	for _, id := range snippetIDs {
		selected[id] = struct{}{}
	}
	requireSelection := len(selected) > 0

	anchors := make([]Anchor, 0, len(artifact.Snippets))
	for _, snippet := range artifact.Snippets {
		if requireSelection {
			if _, ok := selected[snippet.SnippetID]; !ok {
				continue
			}
		}
		expr, err := parser.ParseString(snippet.ExpandedExpression)
		if err != nil {
			return nil, fmt.Errorf("parse snippet %q expression: %w", snippet.SnippetID, err)
		}
		anchors = append(anchors, Anchor{
			Name: snippet.SnippetID,
			Expr: expr,
			Provenance: &AnchorProvenance{
				SourceKind:     AnchorSourceSnippet,
				TargetID:       artifact.TargetID,
				TargetName:     artifact.TargetName,
				ParentTargetID: artifact.Parent.TargetID,
				SnippetID:      snippet.SnippetID,
				PreorderIndex:  snippet.PreorderIndex,
				TreePath:       snippet.TreePath,
				CanonicalKey:   snippet.CanonicalKey,
			},
		})
	}
	if len(anchors) == 0 {
		return nil, fmt.Errorf("no matching snippet anchors selected for target %q", artifact.TargetID)
	}
	return anchors, nil
}

// MazeRealSearchFromSnippetArtifact seeds maze search from snippet-derived
// anchors and uses the parent whole-sample bundle as the target dataset.
func MazeRealSearchFromSnippetArtifact(artifact family.SnippetDatasetArtifact, backend eval.Backend[complex128], snippetIDs []string, options MazeOptions) (MazeReport, error) {
	anchors, err := AnchorsFromSnippetArtifact(artifact, snippetIDs...)
	if err != nil {
		return MazeReport{}, err
	}
	fixture, err := benchmarkCaseFromSnippetArtifact(artifact)
	if err != nil {
		return MazeReport{}, err
	}
	return MazeRealSearch(fixture, backend, anchors, options)
}

// MatchSnippetAnchors compares a target trace against committed snippet traces
// and promotes the best whole-trace matches into explicit maze anchors.
func MatchSnippetAnchors(target common.SearchTarget[float64], artifacts []family.SnippetDatasetArtifact, options SpawnOptions) (SpawnReport, error) {
	if options.TopK <= 0 {
		options.TopK = 3
	}
	if options.MaxScore == 0 {
		options.MaxScore = 1e-6
	}
	variables := target.VariableNames()
	if len(variables) != 1 {
		return SpawnReport{}, fmt.Errorf("snippet matching requires exactly one target variable, got %d", len(variables))
	}
	targetTrace := normalizedTargetTrace(target.Samples(), variables[0])
	if len(targetTrace) == 0 {
		return SpawnReport{}, fmt.Errorf("snippet matching requires non-empty target samples")
	}

	report := SpawnReport{
		Diagnostics: SpawnDiagnostics{
			ArtifactsExamined: len(artifacts),
			BestScore:         math.Inf(1),
		},
	}

	for _, artifact := range artifacts {
		for _, snippet := range artifact.Snippets {
			for _, sampleSet := range artifact.SnippetSamples {
				if sampleSet.SnippetID != snippet.SnippetID {
					continue
				}
				report.Diagnostics.SnippetTracesCompared++
				score, ok := traceMSE(targetTrace, sortPoints(sampleSet.Points))
				if !ok {
					report.Diagnostics.ShapeRejects++
					continue
				}
				match := SnippetMatch{
					TargetID:       artifact.TargetID,
					TargetName:     artifact.TargetName,
					ParentTargetID: artifact.Parent.TargetID,
					SnippetID:      snippet.SnippetID,
					CanonicalKey:   snippet.CanonicalKey,
					DomainID:       sampleSet.DomainID,
					SampleSetID:    sampleSet.SampleSetID,
					Score:          score,
					Provenance: AnchorProvenance{
						SourceKind:     AnchorSourceSnippet,
						TargetID:       artifact.TargetID,
						TargetName:     artifact.TargetName,
						ParentTargetID: artifact.Parent.TargetID,
						SnippetID:      snippet.SnippetID,
						PreorderIndex:  snippet.PreorderIndex,
						TreePath:       snippet.TreePath,
						CanonicalKey:   snippet.CanonicalKey,
					},
				}
				report.Matches = append(report.Matches, match)
				if score < report.Diagnostics.BestScore {
					report.Diagnostics.BestScore = score
				}
			}
		}
	}

	sort.Slice(report.Matches, func(i, j int) bool {
		if report.Matches[i].Score == report.Matches[j].Score {
			if report.Matches[i].CanonicalKey == report.Matches[j].CanonicalKey {
				return report.Matches[i].SnippetID < report.Matches[j].SnippetID
			}
			return report.Matches[i].CanonicalKey < report.Matches[j].CanonicalKey
		}
		return report.Matches[i].Score < report.Matches[j].Score
	})

	for _, match := range report.Matches {
		if match.Score > options.MaxScore {
			report.Diagnostics.ThresholdRejects++
			continue
		}
		if len(report.Anchors) >= options.TopK {
			break
		}
		_, snippet, err := findSnippetByID(artifacts, match.TargetID, match.SnippetID)
		if err != nil {
			return SpawnReport{}, err
		}
		expr, err := parser.ParseString(snippet.ExpandedExpression)
		if err != nil {
			return SpawnReport{}, fmt.Errorf("parse matched snippet %q: %w", snippet.SnippetID, err)
		}
		report.Anchors = append(report.Anchors, Anchor{
			Name:       snippet.SnippetID,
			Expr:       expr,
			Provenance: cloneAnchorProvenance(&match.Provenance),
		})
		report.Diagnostics.AnchorsPromoted++
	}
	if math.IsInf(report.Diagnostics.BestScore, 1) {
		report.Diagnostics.BestScore = 0
	}
	return report, nil
}

// MazeRealSearchFromSpawnedSnippets matches snippet traces against a target,
// promotes the best matches into anchors, and then runs maze from those anchors.
func MazeRealSearchFromSpawnedSnippets(target common.SearchTarget[float64], backend eval.Backend[complex128], artifacts []family.SnippetDatasetArtifact, spawnOptions SpawnOptions, options MazeOptions) (MazeReport, SpawnReport, error) {
	spawnReport, err := MatchSnippetAnchors(target, artifacts, spawnOptions)
	if err != nil {
		return MazeReport{}, SpawnReport{}, err
	}
	if len(spawnReport.Anchors) == 0 {
		return MazeReport{}, spawnReport, fmt.Errorf("no snippet anchors matched within spawn threshold")
	}
	variables := target.VariableNames()
	if len(variables) != 1 {
		return MazeReport{}, spawnReport, fmt.Errorf("maze spawn requires exactly one target variable, got %d", len(variables))
	}
	fixture := common.BenchmarkCase[float64]{
		Name:      "spawned_snippet_target",
		Expr:      nil,
		Samples:   append([]common.Sample[float64](nil), target.Samples()...),
		TargetKey: variables[0],
	}
	report, err := MazeRealSearch(fixture, backend, spawnReport.Anchors, options)
	return report, spawnReport, err
}

// SearchTargetFromSnippetTrace builds a target from one snippet trace inside an artifact.
func SearchTargetFromSnippetTrace(artifact family.SnippetDatasetArtifact, snippetID, domainID, sampleSetID string) (common.StaticTarget[float64], error) {
	var variable string
	for _, domain := range artifact.SamplingDomains {
		if domainID == "" || domain.DomainID == domainID {
			domainID = domain.DomainID
			variable = domain.Sampling.Variable
			break
		}
	}
	if variable == "" {
		return common.StaticTarget[float64]{}, fmt.Errorf("snippet artifact %q has no matching domain", artifact.TargetID)
	}
	for _, sampleSet := range artifact.SnippetSamples {
		if sampleSet.SnippetID != snippetID {
			continue
		}
		if sampleSet.DomainID != domainID {
			continue
		}
		if sampleSetID != "" && sampleSet.SampleSetID != sampleSetID {
			continue
		}
		samples := make([]common.Sample[float64], 0, len(sampleSet.Points))
		for _, point := range sampleSet.Points {
			samples = append(samples, common.Sample[float64]{
				Vars:   map[string]float64{variable: point[0]},
				Target: point[1],
			})
		}
		return common.NewSearchTarget([]string{variable}, samples), nil
	}
	return common.StaticTarget[float64]{}, fmt.Errorf("snippet trace not found for %q in artifact %q", snippetID, artifact.TargetID)
}

func benchmarkCaseFromSnippetArtifact(artifact family.SnippetDatasetArtifact) (common.BenchmarkCase[float64], error) {
	if len(artifact.SamplingDomains) == 0 {
		return common.BenchmarkCase[float64]{}, fmt.Errorf("snippet artifact %q has no sampling domains", artifact.TargetID)
	}
	if len(artifact.WholeSamples) == 0 {
		return common.BenchmarkCase[float64]{}, fmt.Errorf("snippet artifact %q has no whole samples", artifact.TargetID)
	}
	var domainID string
	var variable string
	for i, domain := range artifact.SamplingDomains {
		if i == 0 {
			domainID = domain.DomainID
			variable = domain.Sampling.Variable
			break
		}
	}
	if variable == "" {
		return common.BenchmarkCase[float64]{}, fmt.Errorf("snippet artifact %q has empty sampling variable", artifact.TargetID)
	}

	samples := make([]common.Sample[float64], 0)
	for _, sampleSet := range artifact.WholeSamples {
		if sampleSet.DomainID != domainID {
			continue
		}
		for _, point := range sampleSet.Points {
			samples = append(samples, common.Sample[float64]{
				Vars:   map[string]float64{variable: point[0]},
				Target: point[1],
			})
		}
	}
	if len(samples) == 0 {
		return common.BenchmarkCase[float64]{}, fmt.Errorf("snippet artifact %q has no whole samples for domain %q", artifact.TargetID, domainID)
	}

	return common.BenchmarkCase[float64]{
		Name:      artifact.TargetID,
		Expr:      nil,
		Samples:   samples,
		TargetKey: variable,
	}, nil
}

func normalizedTargetTrace(samples []common.Sample[float64], variable string) [][2]float64 {
	points := make([][2]float64, 0, len(samples))
	for _, sample := range samples {
		x, ok := sample.Vars[variable]
		if !ok {
			continue
		}
		points = append(points, [2]float64{x, sample.Target})
	}
	return sortPoints(points)
}

func sortPoints(points [][2]float64) [][2]float64 {
	out := append([][2]float64(nil), points...)
	sort.Slice(out, func(i, j int) bool {
		if out[i][0] == out[j][0] {
			return out[i][1] < out[j][1]
		}
		return out[i][0] < out[j][0]
	})
	return out
}

func traceMSE(target, candidate [][2]float64) (float64, bool) {
	if len(target) == 0 || len(target) != len(candidate) {
		return 0, false
	}
	var total float64
	for i := range target {
		if math.Abs(target[i][0]-candidate[i][0]) > 1e-9 {
			return 0, false
		}
		diff := target[i][1] - candidate[i][1]
		total += diff * diff
	}
	return total / float64(len(target)), true
}

func findSnippetByID(artifacts []family.SnippetDatasetArtifact, targetID, snippetID string) (family.SnippetDatasetArtifact, family.SnippetDescriptor, error) {
	for _, artifact := range artifacts {
		if artifact.TargetID != targetID {
			continue
		}
		for _, snippet := range artifact.Snippets {
			if snippet.SnippetID == snippetID {
				return artifact, snippet, nil
			}
		}
	}
	return family.SnippetDatasetArtifact{}, family.SnippetDescriptor{}, fmt.Errorf("snippet %q not found in target %q", snippetID, targetID)
}
