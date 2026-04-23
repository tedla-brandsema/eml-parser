package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	TargetKindConcept              = "concept"
	TargetKindRaw                  = "raw"
	DatasetModeRealGrid            = "real_grid"
	DatasetModeExplicitPoints      = "real_points"
	SearchModeEnumerativeReal      = "enumerative_real"
	RecoveryClassExactNormalized   = "exact_normalized_recovery"
	RecoveryClassConceptEquivalent = "concept_equivalent_recovery"
	RecoveryClassApproximateOnly   = "approximate_only_recovery"
	RecoveryClassNoRecovery        = "no_recovery"
)

// Spec is a declarative oracle experiment definition loaded from JSON.
type Spec struct {
	ID          string       `json:"id"`
	Description string       `json:"description"`
	Target      TargetSpec   `json:"target"`
	Dataset     DatasetSpec  `json:"dataset"`
	Search      SearchSpec   `json:"search"`
	Recovery    RecoverySpec `json:"recovery"`
}

// TargetSpec identifies the target law for an experiment.
type TargetSpec struct {
	Kind    string `json:"kind"`
	Concept string `json:"concept,omitempty"`
	RawEML  string `json:"raw_eml,omitempty"`
}

// DatasetSpec describes how real-valued oracle samples are chosen.
type DatasetSpec struct {
	Mode     string    `json:"mode"`
	Variable string    `json:"variable"`
	Grid     *RealGrid `json:"grid,omitempty"`
	Points   []float64 `json:"points,omitempty"`
}

// RealGrid defines an evenly spaced real-valued sampling interval.
type RealGrid struct {
	Start float64 `json:"start"`
	Stop  float64 `json:"stop"`
	Count int     `json:"count"`
}

// SearchSpec configures the current enumerative real search loop.
type SearchSpec struct {
	Mode   string     `json:"mode"`
	Bounds BoundsSpec `json:"bounds"`
	TopN   int        `json:"top_n"`
}

// BoundsSpec mirrors the current bounded raw-tree search options.
type BoundsSpec struct {
	MaxDepth int `json:"max_depth"`
	MaxNodes int `json:"max_nodes"`
}

// RecoverySpec defines the expected recovery criterion for the experiment.
type RecoverySpec struct {
	ExpectedClass         string   `json:"expected_class"`
	ExpectedCanonicalKey  string   `json:"expected_canonical_key,omitempty"`
	AllowedEquivalentKeys []string `json:"allowed_equivalent_keys,omitempty"`
	ApproximateThreshold  *float64 `json:"approximate_threshold,omitempty"`
}

// LoadSpec reads, decodes, and validates a JSON experiment spec.
func LoadSpec(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("read experiment spec: %w", err)
	}
	return ParseSpec(data)
}

// ParseSpec decodes and validates a JSON experiment spec.
func ParseSpec(data []byte) (Spec, error) {
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("decode experiment spec: %w", err)
	}
	if err := spec.Validate(); err != nil {
		return Spec{}, err
	}
	return spec, nil
}

// Validate checks that the spec is compatible with the current experiment
// pipeline.
func (s Spec) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("experiment spec id cannot be empty")
	}
	if strings.TrimSpace(s.Description) == "" {
		return fmt.Errorf("experiment spec %q description cannot be empty", s.ID)
	}
	if err := s.Target.Validate(s.ID); err != nil {
		return err
	}
	if err := s.Dataset.Validate(s.ID); err != nil {
		return err
	}
	if err := s.Search.Validate(s.ID); err != nil {
		return err
	}
	if err := s.Recovery.Validate(s.ID); err != nil {
		return err
	}
	return nil
}

func (t TargetSpec) Validate(specID string) error {
	switch t.Kind {
	case TargetKindConcept:
		if strings.TrimSpace(t.Concept) == "" {
			return fmt.Errorf("experiment spec %q concept target requires concept", specID)
		}
		if strings.TrimSpace(t.RawEML) != "" {
			return fmt.Errorf("experiment spec %q concept target must not set raw_eml", specID)
		}
	case TargetKindRaw:
		if strings.TrimSpace(t.RawEML) == "" {
			return fmt.Errorf("experiment spec %q raw target requires raw_eml", specID)
		}
		if strings.TrimSpace(t.Concept) != "" {
			return fmt.Errorf("experiment spec %q raw target must not set concept", specID)
		}
	default:
		return fmt.Errorf("experiment spec %q target kind must be %q or %q", specID, TargetKindConcept, TargetKindRaw)
	}
	return nil
}

func (d DatasetSpec) Validate(specID string) error {
	if strings.TrimSpace(d.Variable) == "" {
		return fmt.Errorf("experiment spec %q dataset variable cannot be empty", specID)
	}
	switch d.Mode {
	case DatasetModeRealGrid:
		if d.Grid == nil {
			return fmt.Errorf("experiment spec %q real_grid dataset requires grid", specID)
		}
		if len(d.Points) != 0 {
			return fmt.Errorf("experiment spec %q real_grid dataset must not set points", specID)
		}
		if d.Grid.Count <= 0 {
			return fmt.Errorf("experiment spec %q grid count must be positive", specID)
		}
	case DatasetModeExplicitPoints:
		if d.Grid != nil {
			return fmt.Errorf("experiment spec %q real_points dataset must not set grid", specID)
		}
		if len(d.Points) == 0 {
			return fmt.Errorf("experiment spec %q real_points dataset requires points", specID)
		}
	default:
		return fmt.Errorf("experiment spec %q dataset mode must be %q or %q", specID, DatasetModeRealGrid, DatasetModeExplicitPoints)
	}
	return nil
}

func (s SearchSpec) Validate(specID string) error {
	if s.Mode != SearchModeEnumerativeReal {
		return fmt.Errorf("experiment spec %q search mode must be %q", specID, SearchModeEnumerativeReal)
	}
	if s.Bounds.MaxDepth <= 0 {
		return fmt.Errorf("experiment spec %q max_depth must be positive", specID)
	}
	if s.Bounds.MaxNodes <= 0 {
		return fmt.Errorf("experiment spec %q max_nodes must be positive", specID)
	}
	if s.TopN <= 0 {
		return fmt.Errorf("experiment spec %q top_n must be positive", specID)
	}
	return nil
}

func (r RecoverySpec) Validate(specID string) error {
	switch r.ExpectedClass {
	case RecoveryClassExactNormalized:
		if strings.TrimSpace(r.ExpectedCanonicalKey) == "" {
			return fmt.Errorf("experiment spec %q exact recovery requires expected_canonical_key", specID)
		}
	case RecoveryClassConceptEquivalent:
		if strings.TrimSpace(r.ExpectedCanonicalKey) == "" {
			return fmt.Errorf("experiment spec %q concept-equivalent recovery requires expected_canonical_key", specID)
		}
		if len(r.AllowedEquivalentKeys) == 0 {
			return fmt.Errorf("experiment spec %q concept-equivalent recovery requires allowed_equivalent_keys", specID)
		}
	case RecoveryClassApproximateOnly:
		if r.ApproximateThreshold == nil {
			return fmt.Errorf("experiment spec %q approximate-only recovery requires approximate_threshold", specID)
		}
	case RecoveryClassNoRecovery:
		// No additional fields required.
	default:
		return fmt.Errorf(
			"experiment spec %q recovery class must be %q, %q, %q, or %q",
			specID,
			RecoveryClassExactNormalized,
			RecoveryClassConceptEquivalent,
			RecoveryClassApproximateOnly,
			RecoveryClassNoRecovery,
		)
	}
	return nil
}
