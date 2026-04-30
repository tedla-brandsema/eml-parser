package family

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"eml-parser/concepts"
)

const PairedDatasetFormatVersion = "equivalence-paired-v1"

type SamplingDomain struct {
	DomainID string       `json:"domain_id"`
	Sampling SamplingSpec `json:"sampling"`
}

type GroupMemberDescriptor struct {
	Name         string       `json:"name"`
	IsAnchor     bool         `json:"is_anchor"`
	RelationType RelationType `json:"relation_type,omitempty"`
	SourceKind   string       `json:"source_kind"`
	SourceRef    string       `json:"source_ref"`
	CanonicalKey string       `json:"canonical_key"`
}

type EquivalenceGroup struct {
	GroupID      string                  `json:"group_id"`
	DomainID     string                  `json:"domain_id"`
	FamilyID     int                     `json:"family_id"`
	Oracle       []float64               `json:"oracle"`
	Points       [][2]float64            `json:"points"`
	MemberLabels []string                `json:"member_labels"`
	Members      []GroupMemberDescriptor `json:"members"`
}

type PairedDatasetArtifact struct {
	FormatVersion   string             `json:"format_version"`
	DatasetID       string             `json:"dataset_id"`
	FamilyID        int                `json:"family_id"`
	FamilyName      string             `json:"family_name"`
	SamplingDomains []SamplingDomain   `json:"sampling_domains"`
	Groups          []EquivalenceGroup `json:"groups"`
	Notes           string             `json:"notes,omitempty"`
}

type PairedDatasetOptions struct {
	ProjectRoot  string
	Registry     *concepts.Registry
	BaseSampling SamplingSpec
	Families     []EquivalenceFamilySpec
}

func WriteCuratedPairedDatasets(projectRoot string) ([]string, []PairedDatasetArtifact, error) {
	return WritePairedDatasets(PairedDatasetOptions{
		ProjectRoot:  projectRoot,
		Registry:     concepts.StandardLibrary(),
		BaseSampling: DefaultSampling(),
		Families:     CuratedEquivalenceFamilies(),
	})
}

func WritePairedDatasets(opts PairedDatasetOptions) ([]string, []PairedDatasetArtifact, error) {
	if opts.ProjectRoot == "" {
		return nil, nil, fmt.Errorf("project root cannot be empty")
	}
	if opts.Registry == nil {
		opts.Registry = concepts.StandardLibrary()
	}
	if len(opts.Families) == 0 {
		return nil, nil, fmt.Errorf("at least one equivalence family is required")
	}
	if err := validateSampling(opts.BaseSampling); err != nil {
		return nil, nil, err
	}
	if err := validateEquivalenceFamilies(opts.Families); err != nil {
		return nil, nil, err
	}

	outDir := filepath.Join(opts.ProjectRoot, "artifacts", "equivalence", "paired")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create paired output directory: %w", err)
	}

	paths := make([]string, 0, len(opts.Families))
	artifacts := make([]PairedDatasetArtifact, 0, len(opts.Families))
	for _, spec := range opts.Families {
		artifact, err := GeneratePairedDataset(spec, opts.Registry, pairedSamplingDomains(opts.BaseSampling), len(opts.Families))
		if err != nil {
			return nil, nil, err
		}
		path := filepath.Join(outDir, sanitizeFilename(spec.FamilyName)+".paired.json")
		payload, err := json.MarshalIndent(artifact, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("marshal paired dataset %q: %w", spec.FamilyName, err)
		}
		if err := os.WriteFile(path, append(payload, '\n'), 0o644); err != nil {
			return nil, nil, fmt.Errorf("write paired dataset %q: %w", spec.FamilyName, err)
		}
		paths = append(paths, path)
		artifacts = append(artifacts, artifact)
	}
	return paths, artifacts, nil
}

func GeneratePairedDataset(spec EquivalenceFamilySpec, registry *concepts.Registry, domains []SamplingDomain, nFamilies int) (PairedDatasetArtifact, error) {
	if registry == nil {
		registry = concepts.StandardLibrary()
	}
	if err := validateEquivalenceFamilies([]EquivalenceFamilySpec{spec}); err != nil {
		return PairedDatasetArtifact{}, err
	}
	if nFamilies <= 0 {
		return PairedDatasetArtifact{}, fmt.Errorf("nFamilies must be positive")
	}
	if len(domains) == 0 {
		return PairedDatasetArtifact{}, fmt.Errorf("at least one sampling domain is required")
	}
	for _, domain := range domains {
		if domain.DomainID == "" {
			return PairedDatasetArtifact{}, fmt.Errorf("sampling domain id cannot be empty")
		}
		if err := validateSampling(domain.Sampling); err != nil {
			return PairedDatasetArtifact{}, fmt.Errorf("invalid sampling domain %q: %w", domain.DomainID, err)
		}
	}

	groups := make([]EquivalenceGroup, 0)
	samplingDomains := make([]SamplingDomain, 0, len(domains))
	for _, domain := range domains {
		artifact, err := GenerateEquivalenceFamily(spec, registry, domain.Sampling)
		if err != nil {
			return PairedDatasetArtifact{}, fmt.Errorf("generate family %q for domain %q: %w", spec.FamilyName, domain.DomainID, err)
		}
		samplingDomains = append(samplingDomains, domain)
		memberDescriptors := groupMembersFromArtifact(artifact)
		memberLabels := memberLabels(memberDescriptors)
		for _, sampleSet := range artifact.SharedSamples {
			groups = append(groups, EquivalenceGroup{
				GroupID:      domain.DomainID + "_" + sampleSet.SampleSetID,
				DomainID:     domain.DomainID,
				FamilyID:     artifact.FamilyID,
				Oracle:       oneHot(artifact.FamilyID, nFamilies),
				Points:       sampleSet.Points,
				MemberLabels: memberLabels,
				Members:      memberDescriptors,
			})
		}
	}

	return PairedDatasetArtifact{
		FormatVersion:   PairedDatasetFormatVersion,
		DatasetID:       spec.FamilyName + "_paired",
		FamilyID:        spec.FamilyID,
		FamilyName:      spec.FamilyName,
		SamplingDomains: samplingDomains,
		Groups:          groups,
		Notes:           spec.Notes,
	}, nil
}

func pairedSamplingDomains(base SamplingSpec) []SamplingDomain {
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
				Start:       base.Start - 0.25*width,
				Stop:        base.Stop + 0.25*width,
				PointCount:  base.PointCount,
				SampleCount: base.SampleCount,
				Seed:        base.Seed + 1,
			},
		},
	}
}

func groupMembersFromArtifact(artifact EquivalenceFamilyArtifact) []GroupMemberDescriptor {
	out := make([]GroupMemberDescriptor, 0, 1+len(artifact.Members))
	out = append(out, GroupMemberDescriptor{
		Name:         artifact.Anchor.Name,
		IsAnchor:     true,
		SourceKind:   artifact.Anchor.SourceKind,
		SourceRef:    artifact.Anchor.SourceRef,
		CanonicalKey: artifact.Anchor.CanonicalKey,
	})
	for _, member := range artifact.Members {
		out = append(out, GroupMemberDescriptor{
			Name:         member.Name,
			IsAnchor:     false,
			RelationType: member.RelationType,
			SourceKind:   member.SourceKind,
			SourceRef:    member.SourceRef,
			CanonicalKey: member.CanonicalKey,
		})
	}
	return out
}

func memberLabels(members []GroupMemberDescriptor) []string {
	out := make([]string, 0, len(members))
	for _, member := range members {
		out = append(out, member.Name)
	}
	return out
}
