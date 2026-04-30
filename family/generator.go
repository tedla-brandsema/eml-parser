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
	"eml-parser/parser"
	"eml-parser/search"
)

const (
	SourceKindRaw     = "raw"
	SourceKindConcept = "concept"

	GeneratorVersion = "family-artifact-v1"
	realTolerance    = 1e-9
)

type Seed struct {
	FamilyID   int
	FamilyName string
	SourceKind string
	SourceRef  string
}

type SamplingSpec struct {
	Variable    string  `json:"variable"`
	Start       float64 `json:"start"`
	Stop        float64 `json:"stop"`
	PointCount  int     `json:"point_count"`
	SampleCount int     `json:"sample_count"`
	Seed        int64   `json:"seed"`
}

type ConceptProvenance struct {
	Concept                string   `json:"concept"`
	DirectDependencies     []string `json:"direct_dependencies,omitempty"`
	TransitiveDependencies []string `json:"transitive_dependencies,omitempty"`
}

type Sample struct {
	Points   [][2]float64 `json:"points"`
	FamilyID int          `json:"family_id"`
	Oracle   []float64    `json:"oracle"`
}

type Artifact struct {
	GeneratorVersion string             `json:"generator_version"`
	FamilyID         int                `json:"family_id"`
	FamilyName       string             `json:"family_name"`
	SourceKind       string             `json:"source_kind"`
	SourceRef        string             `json:"source_ref"`
	AnchorExpression string             `json:"anchor_expression"`
	CanonicalKey     string             `json:"canonical_key"`
	NFamilies        int                `json:"n_families"`
	Sampling         SamplingSpec       `json:"sampling"`
	Concept          *ConceptProvenance `json:"concept_provenance,omitempty"`
	Samples          []Sample           `json:"samples"`
}

type Options struct {
	ProjectRoot string
	Registry    *concepts.Registry
	Sampling    SamplingSpec
	Seeds       []Seed
}

func CuratedSeeds() []Seed {
	return []Seed{
		{FamilyID: 0, FamilyName: "raw_identity", SourceKind: SourceKindRaw, SourceRef: "x"},
		{FamilyID: 1, FamilyName: "concept_exp", SourceKind: SourceKindConcept, SourceRef: "exp"},
		{FamilyID: 2, FamilyName: "raw_exp_exp", SourceKind: SourceKindRaw, SourceRef: "eml(eml(x, 1), 1)"},
		{FamilyID: 3, FamilyName: "concept_sigmoid", SourceKind: SourceKindConcept, SourceRef: "sigmoid"},
	}
}

func DefaultSampling() SamplingSpec {
	return SamplingSpec{
		Variable:    "x",
		Start:       -0.75,
		Stop:        0.75,
		PointCount:  16,
		SampleCount: 256,
		Seed:        0,
	}
}

func WriteCuratedArtifacts(projectRoot string) ([]string, []Artifact, error) {
	return WriteArtifacts(Options{
		ProjectRoot: projectRoot,
		Registry:    concepts.StandardLibrary(),
		Sampling:    DefaultSampling(),
		Seeds:       CuratedSeeds(),
	})
}

func WriteArtifacts(opts Options) ([]string, []Artifact, error) {
	if opts.ProjectRoot == "" {
		return nil, nil, fmt.Errorf("project root cannot be empty")
	}
	if opts.Registry == nil {
		opts.Registry = concepts.StandardLibrary()
	}
	if len(opts.Seeds) == 0 {
		return nil, nil, fmt.Errorf("at least one family seed is required")
	}
	if err := validateSampling(opts.Sampling); err != nil {
		return nil, nil, err
	}
	if err := validateSeeds(opts.Seeds); err != nil {
		return nil, nil, err
	}

	outDir := filepath.Join(opts.ProjectRoot, "artifacts", "equivalence")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create output directory: %w", err)
	}

	paths := make([]string, 0, len(opts.Seeds))
	artifacts := make([]Artifact, 0, len(opts.Seeds))
	for _, seed := range opts.Seeds {
		artifact, err := GenerateArtifact(seed, opts.Registry, opts.Sampling, len(opts.Seeds))
		if err != nil {
			return nil, nil, err
		}
		path := filepath.Join(outDir, sanitizeFilename(seed.FamilyName)+".json")
		payload, err := json.MarshalIndent(artifact, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("marshal artifact %q: %w", seed.FamilyName, err)
		}
		if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
			return nil, nil, fmt.Errorf("write artifact %q: %w", seed.FamilyName, err)
		}
		paths = append(paths, path)
		artifacts = append(artifacts, artifact)
	}
	return paths, artifacts, nil
}

func GenerateArtifact(seed Seed, registry *concepts.Registry, sampling SamplingSpec, nFamilies int) (Artifact, error) {
	if registry == nil {
		registry = concepts.StandardLibrary()
	}
	if err := validateSampling(sampling); err != nil {
		return Artifact{}, err
	}
	if err := validateSeeds([]Seed{seed}); err != nil {
		return Artifact{}, err
	}
	if nFamilies <= 0 {
		return Artifact{}, fmt.Errorf("nFamilies must be positive")
	}

	expr, conceptProv, err := resolveSeed(seed, registry)
	if err != nil {
		return Artifact{}, err
	}
	candidate := search.NewCandidate(expr)
	samples, err := sampleFamily(candidate.Normalized, seed.FamilyID, nFamilies, sampling)
	if err != nil {
		return Artifact{}, err
	}

	return Artifact{
		GeneratorVersion: GeneratorVersion,
		FamilyID:         seed.FamilyID,
		FamilyName:       seed.FamilyName,
		SourceKind:       seed.SourceKind,
		SourceRef:        seed.SourceRef,
		AnchorExpression: candidate.Normalized.String(),
		CanonicalKey:     candidate.Key,
		NFamilies:        nFamilies,
		Sampling:         sampling,
		Concept:          conceptProv,
		Samples:          samples,
	}, nil
}

func resolveSeed(seed Seed, registry *concepts.Registry) (ast.Expr, *ConceptProvenance, error) {
	switch seed.SourceKind {
	case SourceKindRaw:
		expr, err := parser.ParseString(seed.SourceRef)
		if err != nil {
			return nil, nil, fmt.Errorf("parse raw seed %q: %w", seed.SourceRef, err)
		}
		return expr, nil, nil
	case SourceKindConcept:
		expr, err := registry.ExpandSymbolic(seed.SourceRef)
		if err != nil {
			return nil, nil, fmt.Errorf("expand concept seed %q: %w", seed.SourceRef, err)
		}
		direct, err := registry.DirectDependencies(seed.SourceRef)
		if err != nil {
			return nil, nil, err
		}
		transitive, err := registry.TransitiveDependencies(seed.SourceRef)
		if err != nil {
			return nil, nil, err
		}
		return expr, &ConceptProvenance{
			Concept:                seed.SourceRef,
			DirectDependencies:     direct,
			TransitiveDependencies: transitive,
		}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported source kind %q", seed.SourceKind)
	}
}

func sampleFamily(expr ast.Expr, familyID, nFamilies int, sampling SamplingSpec) ([]Sample, error) {
	rng := rand.New(rand.NewSource(sampling.Seed + int64(familyID)*1000))
	samples := make([]Sample, 0, sampling.SampleCount)
	width := (sampling.Stop - sampling.Start) * 0.6
	if width <= 0 {
		return nil, fmt.Errorf("sampling width must be positive")
	}
	maxStart := sampling.Stop - width
	for sampleIndex := 0; sampleIndex < sampling.SampleCount; sampleIndex++ {
		start := sampling.Start
		if maxStart > sampling.Start {
			start = sampling.Start + rng.Float64()*(maxStart-sampling.Start)
		}
		stop := start + width
		points := make([][2]float64, 0, sampling.PointCount)
		for _, x := range gridPoints(start, stop, sampling.PointCount) {
			value, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{
				sampling.Variable: complex(x, 0),
			})
			if err != nil {
				return nil, fmt.Errorf("evaluate family %q at %g: %w", expr.String(), x, err)
			}
			if math.Abs(imag(value)) > realTolerance {
				return nil, fmt.Errorf("family %q produced non-real output at %g: %v", expr.String(), x, value)
			}
			points = append(points, [2]float64{x, real(value)})
		}
		rng.Shuffle(len(points), func(i, j int) { points[i], points[j] = points[j], points[i] })
		samples = append(samples, Sample{
			Points:   points,
			FamilyID: familyID,
			Oracle:   oneHot(familyID, nFamilies),
		})
	}
	return samples, nil
}

func validateSampling(s SamplingSpec) error {
	if s.Variable == "" {
		return fmt.Errorf("sampling variable cannot be empty")
	}
	if s.Stop <= s.Start {
		return fmt.Errorf("sampling stop must be greater than start")
	}
	if s.PointCount <= 0 {
		return fmt.Errorf("sampling point count must be positive")
	}
	if s.SampleCount <= 0 {
		return fmt.Errorf("sampling sample count must be positive")
	}
	return nil
}

func validateSeeds(seeds []Seed) error {
	ids := make(map[int]struct{}, len(seeds))
	names := make(map[string]struct{}, len(seeds))
	for _, seed := range seeds {
		if seed.FamilyName == "" {
			return fmt.Errorf("family name cannot be empty")
		}
		if seed.SourceRef == "" {
			return fmt.Errorf("source ref cannot be empty for family %q", seed.FamilyName)
		}
		if seed.SourceKind != SourceKindRaw && seed.SourceKind != SourceKindConcept {
			return fmt.Errorf("unsupported source kind %q", seed.SourceKind)
		}
		if _, exists := ids[seed.FamilyID]; exists {
			return fmt.Errorf("duplicate family id %d", seed.FamilyID)
		}
		if _, exists := names[seed.FamilyName]; exists {
			return fmt.Errorf("duplicate family name %q", seed.FamilyName)
		}
		ids[seed.FamilyID] = struct{}{}
		names[seed.FamilyName] = struct{}{}
	}
	return nil
}

func gridPoints(start, stop float64, count int) []float64 {
	if count == 1 {
		return []float64{start}
	}
	step := (stop - start) / float64(count-1)
	points := make([]float64, 0, count)
	for i := 0; i < count; i++ {
		points = append(points, start+float64(i)*step)
	}
	return points
}

func oneHot(index, size int) []float64 {
	out := make([]float64, size)
	if index >= 0 && index < size {
		out[index] = 1
	}
	return out
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "family"
	}
	return b.String()
}
