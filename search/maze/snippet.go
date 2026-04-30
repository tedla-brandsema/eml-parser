package maze

import (
	"encoding/json"
	"fmt"
	"os"

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
