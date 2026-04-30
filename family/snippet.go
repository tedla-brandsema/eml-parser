package family

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/eval"
	"eml-parser/search"
)

const SnippetDatasetFormatVersion = "snippet-dataset-v1"

type SnippetSelector struct {
	Name          string `json:"name"`
	PreorderIndex int    `json:"preorder_index"`
}

type SnippetTargetSpec struct {
	TargetID   string            `json:"target_id"`
	TargetName string            `json:"target_name"`
	SourceKind string            `json:"source_kind"`
	SourceRef  string            `json:"source_ref"`
	Selectors  []SnippetSelector `json:"selectors"`
	Notes      string            `json:"notes,omitempty"`
}

type SnippetStats struct {
	NodeCount int `json:"node_count"`
	TreeDepth int `json:"tree_depth"`
	LeafCount int `json:"leaf_count"`
}

type SnippetParent struct {
	TargetID             string             `json:"target_id"`
	TargetName           string             `json:"target_name"`
	SourceKind           string             `json:"source_kind"`
	SourceRef            string             `json:"source_ref"`
	ExpandedExpression   string             `json:"expanded_expression"`
	NormalizedExpression string             `json:"normalized_expression"`
	CanonicalKey         string             `json:"canonical_key"`
	Concept              *ConceptProvenance `json:"concept_provenance,omitempty"`
	Stats                SnippetStats       `json:"stats"`
}

type SnippetDescriptor struct {
	SnippetID            string       `json:"snippet_id"`
	Name                 string       `json:"name"`
	PreorderIndex        int          `json:"preorder_index"`
	TreePath             string       `json:"tree_path"`
	ExpandedExpression   string       `json:"expanded_expression"`
	NormalizedExpression string       `json:"normalized_expression"`
	CanonicalKey         string       `json:"canonical_key"`
	Stats                SnippetStats `json:"stats"`
}

type WholeTargetSampleSet struct {
	DomainID    string       `json:"domain_id"`
	SampleSetID string       `json:"sample_set_id"`
	Points      [][2]float64 `json:"points"`
}

type SnippetSampleSet struct {
	SnippetID   string       `json:"snippet_id"`
	DomainID    string       `json:"domain_id"`
	SampleSetID string       `json:"sample_set_id"`
	Points      [][2]float64 `json:"points"`
}

type SnippetDatasetArtifact struct {
	FormatVersion   string                 `json:"format_version"`
	TargetID        string                 `json:"target_id"`
	TargetName      string                 `json:"target_name"`
	Parent          SnippetParent          `json:"parent"`
	Snippets        []SnippetDescriptor    `json:"snippets"`
	SamplingDomains []SamplingDomain       `json:"sampling_domains"`
	WholeSamples    []WholeTargetSampleSet `json:"whole_samples"`
	SnippetSamples  []SnippetSampleSet     `json:"snippet_samples"`
	Notes           string                 `json:"notes,omitempty"`
}

type SnippetDatasetOptions struct {
	ProjectRoot  string
	Registry     *concepts.Registry
	BaseSampling SamplingSpec
	Targets      []SnippetTargetSpec
}

type subtreeRef struct {
	Index int
	Path  string
	Expr  ast.Expr
}

func CuratedSnippetTargets() []SnippetTargetSpec {
	return []SnippetTargetSpec{
		{
			TargetID:   "raw_exp3",
			TargetName: "Raw Exp Depth 3",
			SourceKind: SourceKindRaw,
			SourceRef:  "eml(eml(eml(x, 1), 1), 1)",
			Selectors: []SnippetSelector{
				{Name: "exp2", PreorderIndex: 1},
				{Name: "exp1", PreorderIndex: 2},
				{Name: "x_leaf", PreorderIndex: 3},
			},
			Notes: "Nested exponential chain with overlapping subtree snippets.",
		},
		{
			TargetID:   "raw_exp4",
			TargetName: "Raw Exp Depth 4",
			SourceKind: SourceKindRaw,
			SourceRef:  "eml(eml(eml(eml(x, 1), 1), 1), 1)",
			Selectors: []SnippetSelector{
				{Name: "exp3", PreorderIndex: 1},
				{Name: "exp2", PreorderIndex: 2},
				{Name: "exp1", PreorderIndex: 3},
			},
			Notes: "Deeper overlapping exponential snippets for assembly-oriented tasks.",
		},
		{
			TargetID:   "concept_id",
			TargetName: "Concept Identity",
			SourceKind: SourceKindConcept,
			SourceRef:  "id",
			Selectors: []SnippetSelector{
				{Name: "log_branch", PreorderIndex: 1},
				{Name: "inner_branch", PreorderIndex: 4},
				{Name: "variable_leaf", PreorderIndex: 6},
			},
			Notes: "Concept-derived parent target with nested exact structural snippet labels.",
		},
	}
}

func DefaultSnippetSampling() SamplingSpec {
	return SamplingSpec{
		Variable:    "x",
		Start:       0.05,
		Stop:        0.35,
		PointCount:  16,
		SampleCount: 64,
		Seed:        0,
	}
}

func WriteCuratedSnippetDatasets(projectRoot string) ([]string, []SnippetDatasetArtifact, error) {
	return WriteSnippetDatasets(SnippetDatasetOptions{
		ProjectRoot:  projectRoot,
		Registry:     concepts.StandardLibrary(),
		BaseSampling: DefaultSnippetSampling(),
		Targets:      CuratedSnippetTargets(),
	})
}

func WriteSnippetDatasets(opts SnippetDatasetOptions) ([]string, []SnippetDatasetArtifact, error) {
	if opts.ProjectRoot == "" {
		return nil, nil, fmt.Errorf("project root cannot be empty")
	}
	if opts.Registry == nil {
		opts.Registry = concepts.StandardLibrary()
	}
	if len(opts.Targets) == 0 {
		return nil, nil, fmt.Errorf("at least one snippet target is required")
	}
	if err := validateSampling(opts.BaseSampling); err != nil {
		return nil, nil, err
	}
	if err := validateSnippetTargets(opts.Targets); err != nil {
		return nil, nil, err
	}

	outDir := filepath.Join(opts.ProjectRoot, "artifacts", "snippets")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create snippet output directory: %w", err)
	}

	paths := make([]string, 0, len(opts.Targets))
	artifacts := make([]SnippetDatasetArtifact, 0, len(opts.Targets))
	domains := snippetSamplingDomains(opts.BaseSampling)
	for _, target := range opts.Targets {
		artifact, err := GenerateSnippetDataset(target, opts.Registry, domains)
		if err != nil {
			return nil, nil, err
		}
		path := filepath.Join(outDir, sanitizeFilename(target.TargetID)+".snippet.json")
		payload, err := json.MarshalIndent(artifact, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("marshal snippet dataset %q: %w", target.TargetID, err)
		}
		if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
			return nil, nil, fmt.Errorf("write snippet dataset %q: %w", target.TargetID, err)
		}
		paths = append(paths, path)
		artifacts = append(artifacts, artifact)
	}
	return paths, artifacts, nil
}

func GenerateSnippetDataset(target SnippetTargetSpec, registry *concepts.Registry, domains []SamplingDomain) (SnippetDatasetArtifact, error) {
	if registry == nil {
		registry = concepts.StandardLibrary()
	}
	if err := validateSnippetTargets([]SnippetTargetSpec{target}); err != nil {
		return SnippetDatasetArtifact{}, err
	}
	if len(domains) == 0 {
		return SnippetDatasetArtifact{}, fmt.Errorf("at least one sampling domain is required")
	}
	for _, domain := range domains {
		if domain.DomainID == "" {
			return SnippetDatasetArtifact{}, fmt.Errorf("sampling domain id cannot be empty")
		}
		if err := validateSampling(domain.Sampling); err != nil {
			return SnippetDatasetArtifact{}, fmt.Errorf("invalid sampling domain %q: %w", domain.DomainID, err)
		}
	}

	expr, conceptProv, err := resolveSeed(Seed{
		SourceKind: target.SourceKind,
		SourceRef:  target.SourceRef,
	}, registry)
	if err != nil {
		return SnippetDatasetArtifact{}, fmt.Errorf("resolve snippet target %q: %w", target.TargetID, err)
	}
	parentCandidate := search.NewCandidate(expr)
	parent := SnippetParent{
		TargetID:             target.TargetID,
		TargetName:           target.TargetName,
		SourceKind:           target.SourceKind,
		SourceRef:            target.SourceRef,
		ExpandedExpression:   expr.String(),
		NormalizedExpression: parentCandidate.Normalized.String(),
		CanonicalKey:         parentCandidate.Key,
		Concept:              conceptProv,
		Stats:                snippetStats(expr),
	}

	all := preorderSubtrees(expr)
	snippets := make([]SnippetDescriptor, 0, len(target.Selectors))
	snippetRefs := make([]subtreeRef, 0, len(target.Selectors))
	for _, selector := range target.Selectors {
		if selector.PreorderIndex < 0 || selector.PreorderIndex >= len(all) {
			return SnippetDatasetArtifact{}, fmt.Errorf("snippet selector %q index %d out of range for target %q", selector.Name, selector.PreorderIndex, target.TargetID)
		}
		ref := all[selector.PreorderIndex]
		candidate := search.NewCandidate(ref.Expr)
		snippets = append(snippets, SnippetDescriptor{
			SnippetID:            target.TargetID + "_" + selector.Name,
			Name:                 selector.Name,
			PreorderIndex:        selector.PreorderIndex,
			TreePath:             ref.Path,
			ExpandedExpression:   ref.Expr.String(),
			NormalizedExpression: candidate.Normalized.String(),
			CanonicalKey:         candidate.Key,
			Stats:                snippetStats(ref.Expr),
		})
		snippetRefs = append(snippetRefs, ref)
	}

	wholeSamples := make([]WholeTargetSampleSet, 0)
	snippetSamples := make([]SnippetSampleSet, 0)
	for _, domain := range domains {
		coordinateSets := sampleCoordinates(domain.Sampling)
		parentSampleSets, err := sampleExpressionOnCoordinates(parentCandidate.Normalized, coordinateSets, domain.Sampling.Variable)
		if err != nil {
			return SnippetDatasetArtifact{}, fmt.Errorf("sample parent %q on domain %q: %w", target.TargetID, domain.DomainID, err)
		}
		for _, sampleSet := range parentSampleSets {
			wholeSamples = append(wholeSamples, WholeTargetSampleSet{
				DomainID:    domain.DomainID,
				SampleSetID: sampleSet.SampleSetID,
				Points:      sampleSet.Points,
			})
		}
		for i, snippet := range snippets {
			snippetSetPoints, err := sampleExpressionOnCoordinates(snippetRefs[i].Expr, coordinateSets, domain.Sampling.Variable)
			if err != nil {
				return SnippetDatasetArtifact{}, fmt.Errorf("sample snippet %q on domain %q: %w", snippet.SnippetID, domain.DomainID, err)
			}
			for _, sampleSet := range snippetSetPoints {
				snippetSamples = append(snippetSamples, SnippetSampleSet{
					SnippetID:   snippet.SnippetID,
					DomainID:    domain.DomainID,
					SampleSetID: sampleSet.SampleSetID,
					Points:      sampleSet.Points,
				})
			}
		}
	}

	return SnippetDatasetArtifact{
		FormatVersion:   SnippetDatasetFormatVersion,
		TargetID:        target.TargetID,
		TargetName:      target.TargetName,
		Parent:          parent,
		Snippets:        snippets,
		SamplingDomains: domains,
		WholeSamples:    wholeSamples,
		SnippetSamples:  snippetSamples,
		Notes:           target.Notes,
	}, nil
}

func snippetSamplingDomains(base SamplingSpec) []SamplingDomain {
	width := base.Stop - base.Start
	return []SamplingDomain{
		{
			DomainID: "default",
			Sampling: base,
		},
		{
			DomainID: "widened",
			Sampling: SamplingSpec{
				Variable:    base.Variable,
				Start:       math.Max(0.1, base.Start-0.15*width),
				Stop:        base.Stop + 0.35*width,
				PointCount:  base.PointCount,
				SampleCount: base.SampleCount,
				Seed:        base.Seed + 1,
			},
		},
	}
}

func sampleCoordinates(sampling SamplingSpec) []SharedSampleSet {
	rng := rand.New(rand.NewSource(sampling.Seed))
	width := (sampling.Stop - sampling.Start) * 0.6
	maxStart := sampling.Stop - width
	out := make([]SharedSampleSet, 0, sampling.SampleCount)
	for i := 0; i < sampling.SampleCount; i++ {
		start := sampling.Start
		if maxStart > sampling.Start {
			start = sampling.Start + rng.Float64()*(maxStart-sampling.Start)
		}
		stop := start + width
		points := make([][2]float64, 0, sampling.PointCount)
		for _, x := range gridPoints(start, stop, sampling.PointCount) {
			points = append(points, [2]float64{x, 0})
		}
		rng.Shuffle(len(points), func(i, j int) { points[i], points[j] = points[j], points[i] })
		out = append(out, SharedSampleSet{
			SampleSetID: fmt.Sprintf("sample_%03d", i),
			Points:      points,
		})
	}
	return out
}

func sampleExpressionOnCoordinates(expr ast.Expr, coords []SharedSampleSet, variable string) ([]SharedSampleSet, error) {
	out := make([]SharedSampleSet, 0, len(coords))
	for _, coordSet := range coords {
		points := make([][2]float64, 0, len(coordSet.Points))
		for _, point := range coordSet.Points {
			value, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{
				variable: complex(point[0], 0),
			})
			if err != nil {
				return nil, fmt.Errorf("evaluate %q at %g: %w", expr.String(), point[0], err)
			}
			if math.Abs(imag(value)) > realTolerance {
				return nil, fmt.Errorf("expression %q produced non-real output at %g: %v", expr.String(), point[0], value)
			}
			points = append(points, [2]float64{point[0], real(value)})
		}
		out = append(out, SharedSampleSet{
			SampleSetID: coordSet.SampleSetID,
			Points:      points,
		})
	}
	return out, nil
}

func preorderSubtrees(expr ast.Expr) []subtreeRef {
	var out []subtreeRef
	var walk func(node ast.Expr, path string)
	index := 0
	walk = func(node ast.Expr, path string) {
		out = append(out, subtreeRef{
			Index: index,
			Path:  path,
			Expr:  cloneExpr(node),
		})
		index++
		if app, ok := node.(ast.Apply); ok {
			walk(app.Left, path+".L")
			walk(app.Right, path+".R")
		}
	}
	walk(expr, "root")
	return out
}

func cloneExpr(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case ast.One:
		return ast.One{}
	case ast.Variable:
		return ast.Variable{Name: n.Name}
	case ast.Apply:
		return ast.Apply{
			Left:  cloneExpr(n.Left),
			Right: cloneExpr(n.Right),
		}
	default:
		return nil
	}
}

func snippetStats(expr ast.Expr) SnippetStats {
	stats := search.TreeStats(expr)
	return SnippetStats{
		NodeCount: stats.NodeCount,
		TreeDepth: stats.TreeDepth,
		LeafCount: stats.LeafCount,
	}
}

func validateSnippetTargets(targets []SnippetTargetSpec) error {
	ids := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		if target.TargetID == "" {
			return fmt.Errorf("snippet target id cannot be empty")
		}
		if target.TargetName == "" {
			return fmt.Errorf("snippet target name cannot be empty")
		}
		if target.SourceKind != SourceKindRaw && target.SourceKind != SourceKindConcept {
			return fmt.Errorf("unsupported snippet target source kind %q", target.SourceKind)
		}
		if target.SourceRef == "" {
			return fmt.Errorf("snippet target %q source ref cannot be empty", target.TargetID)
		}
		if _, exists := ids[target.TargetID]; exists {
			return fmt.Errorf("duplicate snippet target id %q", target.TargetID)
		}
		if len(target.Selectors) == 0 {
			return fmt.Errorf("snippet target %q must have at least one selector", target.TargetID)
		}
		names := make(map[string]struct{}, len(target.Selectors))
		indices := make(map[int]struct{}, len(target.Selectors))
		for _, selector := range target.Selectors {
			if selector.Name == "" {
				return fmt.Errorf("snippet selector name cannot be empty for target %q", target.TargetID)
			}
			if selector.PreorderIndex < 0 {
				return fmt.Errorf("snippet selector %q has negative preorder index for target %q", selector.Name, target.TargetID)
			}
			if _, exists := names[selector.Name]; exists {
				return fmt.Errorf("duplicate snippet selector name %q for target %q", selector.Name, target.TargetID)
			}
			if _, exists := indices[selector.PreorderIndex]; exists {
				return fmt.Errorf("duplicate snippet selector index %d for target %q", selector.PreorderIndex, target.TargetID)
			}
			names[selector.Name] = struct{}{}
			indices[selector.PreorderIndex] = struct{}{}
		}
		ids[target.TargetID] = struct{}{}
	}
	return nil
}

func snippetContains(a, b SnippetDescriptor) bool {
	return strings.HasPrefix(b.TreePath, a.TreePath+".") || a.TreePath == b.TreePath
}
