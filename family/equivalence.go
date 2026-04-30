package family

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/eval"
	"eml-parser/search"
)

const EquivalenceFamilyFormatVersion = "equivalence-family-v2"

type RelationType string

const (
	RelationExactSameRawTree         RelationType = "exact_same_raw_tree"
	RelationNormalizedSameRawTree    RelationType = "normalized_same_raw_tree"
	RelationKnownConceptEquivalence  RelationType = "known_concept_equivalence"
	RelationSampledNumericEquivalent RelationType = "sampled_numeric_equivalence"
	RelationContextualSubstitution   RelationType = "contextual_subtree_substitution"
)

type EntrySeed struct {
	Name       string
	SourceKind string
	SourceRef  string
}

type MemberSpec struct {
	Name         string
	RelationType RelationType
	SourceKind   string
	SourceRef    string
}

type EquivalenceFamilySpec struct {
	FamilyID   int
	FamilyName string
	Anchor     EntrySeed
	Members    []MemberSpec
	Notes      string
}

type FamilyEntry struct {
	Name                 string             `json:"name"`
	SourceKind           string             `json:"source_kind"`
	SourceRef            string             `json:"source_ref"`
	ExpandedExpression   string             `json:"expanded_expression"`
	NormalizedExpression string             `json:"normalized_expression"`
	CanonicalKey         string             `json:"canonical_key"`
	Concept              *ConceptProvenance `json:"concept_provenance,omitempty"`
}

type FamilyMember struct {
	RelationType RelationType `json:"relation_type"`
	FamilyEntry
}

type SharedSampleSet struct {
	SampleSetID string       `json:"sample_set_id"`
	Points      [][2]float64 `json:"points"`
}

type EquivalenceFamilyArtifact struct {
	FormatVersion string            `json:"format_version"`
	FamilyID      int               `json:"family_id"`
	FamilyName    string            `json:"family_name"`
	Anchor        FamilyEntry       `json:"anchor"`
	Members       []FamilyMember    `json:"members"`
	Sampling      SamplingSpec      `json:"sampling"`
	SharedSamples []SharedSampleSet `json:"shared_samples"`
	Notes         string            `json:"notes,omitempty"`
}

func CuratedEquivalenceFamilies() []EquivalenceFamilySpec {
	return []EquivalenceFamilySpec{
		{
			FamilyID:   0,
			FamilyName: "identity_exact",
			Anchor: EntrySeed{
				Name:       "raw_identity_anchor",
				SourceKind: SourceKindRaw,
				SourceRef:  "x",
			},
			Members: []MemberSpec{
				{
					Name:         "raw_identity_duplicate",
					RelationType: RelationExactSameRawTree,
					SourceKind:   SourceKindRaw,
					SourceRef:    "x",
				},
			},
			Notes: "Sanity-check family for exact raw equality.",
		},
		{
			FamilyID:   1,
			FamilyName: "identity_normalized",
			Anchor: EntrySeed{
				Name:       "concept_identity_anchor",
				SourceKind: SourceKindConcept,
				SourceRef:  "id",
			},
			Members: []MemberSpec{
				{
					Name:         "raw_identity_member",
					RelationType: RelationNormalizedSameRawTree,
					SourceKind:   SourceKindRaw,
					SourceRef:    "x",
				},
			},
			Notes: "Anchor and member normalize to the same raw EML tree.",
		},
		{
			FamilyID:   2,
			FamilyName: "exp_known_equivalence",
			Anchor: EntrySeed{
				Name:       "concept_exp_anchor",
				SourceKind: SourceKindConcept,
				SourceRef:  "exp",
			},
			Members: []MemberSpec{
				{
					Name:         "raw_exp_member",
					RelationType: RelationKnownConceptEquivalence,
					SourceKind:   SourceKindRaw,
					SourceRef:    "eml(x, 1)",
				},
			},
			Notes: "Concept-level and raw-tree encodings of exp(x).",
		},
	}
}

func WriteCuratedEquivalenceFamilies(projectRoot string) ([]string, []EquivalenceFamilyArtifact, error) {
	return WriteEquivalenceFamilies(EquivalenceOptions{
		ProjectRoot: projectRoot,
		Registry:    concepts.StandardLibrary(),
		Sampling:    DefaultSampling(),
		Families:    CuratedEquivalenceFamilies(),
	})
}

type EquivalenceOptions struct {
	ProjectRoot string
	Registry    *concepts.Registry
	Sampling    SamplingSpec
	Families    []EquivalenceFamilySpec
}

func WriteEquivalenceFamilies(opts EquivalenceOptions) ([]string, []EquivalenceFamilyArtifact, error) {
	if opts.ProjectRoot == "" {
		return nil, nil, fmt.Errorf("project root cannot be empty")
	}
	if opts.Registry == nil {
		opts.Registry = concepts.StandardLibrary()
	}
	if len(opts.Families) == 0 {
		return nil, nil, fmt.Errorf("at least one equivalence family is required")
	}
	if err := validateSampling(opts.Sampling); err != nil {
		return nil, nil, err
	}
	if err := validateEquivalenceFamilies(opts.Families); err != nil {
		return nil, nil, err
	}

	outDir := filepath.Join(opts.ProjectRoot, "artifacts", "equivalence")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create output directory: %w", err)
	}

	paths := make([]string, 0, len(opts.Families))
	artifacts := make([]EquivalenceFamilyArtifact, 0, len(opts.Families))
	for _, spec := range opts.Families {
		artifact, err := GenerateEquivalenceFamily(spec, opts.Registry, opts.Sampling)
		if err != nil {
			return nil, nil, err
		}
		path := filepath.Join(outDir, sanitizeFilename(spec.FamilyName)+".family.json")
		payload, err := json.MarshalIndent(artifact, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("marshal equivalence family %q: %w", spec.FamilyName, err)
		}
		if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
			return nil, nil, fmt.Errorf("write equivalence family %q: %w", spec.FamilyName, err)
		}
		paths = append(paths, path)
		artifacts = append(artifacts, artifact)
	}
	return paths, artifacts, nil
}

func GenerateEquivalenceFamily(spec EquivalenceFamilySpec, registry *concepts.Registry, sampling SamplingSpec) (EquivalenceFamilyArtifact, error) {
	if registry == nil {
		registry = concepts.StandardLibrary()
	}
	if err := validateSampling(sampling); err != nil {
		return EquivalenceFamilyArtifact{}, err
	}
	if err := validateEquivalenceFamilies([]EquivalenceFamilySpec{spec}); err != nil {
		return EquivalenceFamilyArtifact{}, err
	}

	anchorExpr, anchorProv, err := resolveSeed(Seed{
		SourceKind: spec.Anchor.SourceKind,
		SourceRef:  spec.Anchor.SourceRef,
	}, registry)
	if err != nil {
		return EquivalenceFamilyArtifact{}, fmt.Errorf("resolve anchor %q: %w", spec.FamilyName, err)
	}
	anchorEntry := buildFamilyEntry(spec.Anchor.Name, spec.Anchor.SourceKind, spec.Anchor.SourceRef, anchorExpr, anchorProv)
	sampleSets, err := sharedSampleSets(anchorExpr, sampling)
	if err != nil {
		return EquivalenceFamilyArtifact{}, err
	}

	members := make([]FamilyMember, 0, len(spec.Members))
	for _, memberSpec := range spec.Members {
		memberExpr, memberProv, err := resolveSeed(Seed{
			SourceKind: memberSpec.SourceKind,
			SourceRef:  memberSpec.SourceRef,
		}, registry)
		if err != nil {
			return EquivalenceFamilyArtifact{}, fmt.Errorf("resolve member %q in family %q: %w", memberSpec.Name, spec.FamilyName, err)
		}
		memberEntry := buildFamilyEntry(memberSpec.Name, memberSpec.SourceKind, memberSpec.SourceRef, memberExpr, memberProv)
		if err := validateMemberAgreement(spec.FamilyName, memberEntry, memberExpr, sampleSets, sampling.Variable); err != nil {
			return EquivalenceFamilyArtifact{}, err
		}
		members = append(members, FamilyMember{
			RelationType: memberSpec.RelationType,
			FamilyEntry:  memberEntry,
		})
	}

	return EquivalenceFamilyArtifact{
		FormatVersion: EquivalenceFamilyFormatVersion,
		FamilyID:      spec.FamilyID,
		FamilyName:    spec.FamilyName,
		Anchor:        anchorEntry,
		Members:       members,
		Sampling:      sampling,
		SharedSamples: sampleSets,
		Notes:         spec.Notes,
	}, nil
}

func buildFamilyEntry(name, sourceKind, sourceRef string, expr ast.Expr, prov *ConceptProvenance) FamilyEntry {
	candidate := search.NewCandidate(expr)
	return FamilyEntry{
		Name:                 name,
		SourceKind:           sourceKind,
		SourceRef:            sourceRef,
		ExpandedExpression:   expr.String(),
		NormalizedExpression: candidate.Normalized.String(),
		CanonicalKey:         candidate.Key,
		Concept:              prov,
	}
}

func sharedSampleSets(expr ast.Expr, sampling SamplingSpec) ([]SharedSampleSet, error) {
	rng := randSource(sampling.Seed)
	width := (sampling.Stop - sampling.Start) * 0.6
	if width <= 0 {
		return nil, fmt.Errorf("sampling width must be positive")
	}
	maxStart := sampling.Stop - width
	sampleSets := make([]SharedSampleSet, 0, sampling.SampleCount)
	for i := 0; i < sampling.SampleCount; i++ {
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
				return nil, fmt.Errorf("evaluate anchor %q at %g: %w", expr.String(), x, err)
			}
			if math.Abs(imag(value)) > realTolerance {
				return nil, fmt.Errorf("anchor %q produced non-real output at %g: %v", expr.String(), x, value)
			}
			points = append(points, [2]float64{x, real(value)})
		}
		rng.Shuffle(len(points), func(i, j int) { points[i], points[j] = points[j], points[i] })
		sampleSets = append(sampleSets, SharedSampleSet{
			SampleSetID: fmt.Sprintf("sample_%03d", i),
			Points:      points,
		})
	}
	return sampleSets, nil
}

func validateMemberAgreement(familyName string, entry FamilyEntry, expr ast.Expr, sampleSets []SharedSampleSet, variable string) error {
	for _, sampleSet := range sampleSets {
		for _, point := range sampleSet.Points {
			value, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{
				variable: complex(point[0], 0),
			})
			if err != nil {
				return fmt.Errorf("evaluate member %q in family %q at %g: %w", entry.Name, familyName, point[0], err)
			}
			if math.Abs(imag(value)) > realTolerance {
				return fmt.Errorf("member %q in family %q produced non-real output at %g: %v", entry.Name, familyName, point[0], value)
			}
			if math.Abs(real(value)-point[1]) > realTolerance {
				return fmt.Errorf(
					"member %q in family %q disagrees with anchor on %s at %g: got %g want %g",
					entry.Name,
					familyName,
					sampleSet.SampleSetID,
					point[0],
					real(value),
					point[1],
				)
			}
		}
	}
	return nil
}

func validateEquivalenceFamilies(families []EquivalenceFamilySpec) error {
	ids := make(map[int]struct{}, len(families))
	names := make(map[string]struct{}, len(families))
	for _, family := range families {
		if family.FamilyName == "" {
			return fmt.Errorf("equivalence family name cannot be empty")
		}
		if _, exists := ids[family.FamilyID]; exists {
			return fmt.Errorf("duplicate equivalence family id %d", family.FamilyID)
		}
		if _, exists := names[family.FamilyName]; exists {
			return fmt.Errorf("duplicate equivalence family name %q", family.FamilyName)
		}
		if err := validateEntrySeed(family.Anchor); err != nil {
			return fmt.Errorf("invalid anchor in family %q: %w", family.FamilyName, err)
		}
		if len(family.Members) == 0 {
			return fmt.Errorf("equivalence family %q must have at least one member", family.FamilyName)
		}
		memberNames := make(map[string]struct{}, len(family.Members))
		for _, member := range family.Members {
			if member.Name == "" {
				return fmt.Errorf("member name cannot be empty in family %q", family.FamilyName)
			}
			if _, exists := memberNames[member.Name]; exists {
				return fmt.Errorf("duplicate member name %q in family %q", member.Name, family.FamilyName)
			}
			if !isValidRelationType(member.RelationType) {
				return fmt.Errorf("invalid relation type %q in family %q", member.RelationType, family.FamilyName)
			}
			if err := validateEntrySeed(EntrySeed{Name: member.Name, SourceKind: member.SourceKind, SourceRef: member.SourceRef}); err != nil {
				return fmt.Errorf("invalid member %q in family %q: %w", member.Name, family.FamilyName, err)
			}
			memberNames[member.Name] = struct{}{}
		}
		ids[family.FamilyID] = struct{}{}
		names[family.FamilyName] = struct{}{}
	}
	return nil
}

func validateEntrySeed(seed EntrySeed) error {
	if seed.Name == "" {
		return fmt.Errorf("entry name cannot be empty")
	}
	if seed.SourceRef == "" {
		return fmt.Errorf("source ref cannot be empty")
	}
	if seed.SourceKind != SourceKindRaw && seed.SourceKind != SourceKindConcept {
		return fmt.Errorf("unsupported source kind %q", seed.SourceKind)
	}
	return nil
}

func isValidRelationType(rt RelationType) bool {
	switch rt {
	case RelationExactSameRawTree,
		RelationNormalizedSameRawTree,
		RelationKnownConceptEquivalence,
		RelationSampledNumericEquivalent,
		RelationContextualSubstitution:
		return true
	default:
		return false
	}
}

func randSource(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
