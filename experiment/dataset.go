package experiment

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/eval"
	"eml-parser/parser"
	"eml-parser/search"
)

const nonRealTolerance = 1e-9

// DatasetArtifact is the generated real-valued dataset for one oracle
// experiment spec.
type DatasetArtifact struct {
	ExperimentID string          `json:"experiment_id"`
	Description  string          `json:"description"`
	Target       DatasetTarget   `json:"target"`
	Variable     string          `json:"variable"`
	Mode         string          `json:"mode"`
	Domain       DatasetDomain   `json:"domain"`
	SampleCount  int             `json:"sample_count"`
	Samples      []DatasetSample `json:"samples"`
}

// DatasetTarget describes the target law used to generate dataset values.
type DatasetTarget struct {
	Kind         string `json:"kind"`
	Concept      string `json:"concept,omitempty"`
	RawEML       string `json:"raw_eml,omitempty"`
	CanonicalKey string `json:"canonical_key"`
	Expression   string `json:"expression"`
}

// DatasetDomain captures the real-valued sampling domain for the dataset.
type DatasetDomain struct {
	Grid   *RealGrid `json:"grid,omitempty"`
	Points []float64 `json:"points,omitempty"`
}

// DatasetSample is one real-valued oracle observation.
type DatasetSample struct {
	Input  float64 `json:"input"`
	Target float64 `json:"target"`
}

// BuildDataset generates a deterministic real-valued dataset artifact from one
// validated experiment spec.
func BuildDataset(spec Spec) (DatasetArtifact, error) {
	expr, targetInfo, err := resolveTarget(spec)
	if err != nil {
		return DatasetArtifact{}, err
	}

	samplePoints, domain, err := samplePointsForDataset(spec.Dataset)
	if err != nil {
		return DatasetArtifact{}, err
	}

	samples := make([]DatasetSample, 0, len(samplePoints))
	for _, point := range samplePoints {
		value, err := eval.EvaluateMap(expr, eval.Complex128Backend{}, map[string]complex128{
			spec.Dataset.Variable: complex(point, 0),
		})
		if err != nil {
			return DatasetArtifact{}, fmt.Errorf("evaluate target for %q at %g: %w", spec.ID, point, err)
		}
		if math.Abs(imag(value)) > nonRealTolerance {
			return DatasetArtifact{}, fmt.Errorf(
				"target %q produced non-real output at %g: %v",
				spec.ID,
				point,
				value,
			)
		}
		samples = append(samples, DatasetSample{
			Input:  point,
			Target: real(value),
		})
	}

	return DatasetArtifact{
		ExperimentID: spec.ID,
		Description:  spec.Description,
		Target:       targetInfo,
		Variable:     spec.Dataset.Variable,
		Mode:         spec.Dataset.Mode,
		Domain:       domain,
		SampleCount:  len(samples),
		Samples:      samples,
	}, nil
}

// WriteDataset generates and writes one dataset artifact into
// experiments/datasets/<experiment-id>.json under the supplied project root.
func WriteDataset(projectRoot string, spec Spec) (string, DatasetArtifact, error) {
	artifact, err := BuildDataset(spec)
	if err != nil {
		return "", DatasetArtifact{}, err
	}

	outputPath := filepath.Join(projectRoot, "experiments", "datasets", spec.ID+".json")
	payload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return "", DatasetArtifact{}, fmt.Errorf("marshal dataset artifact: %w", err)
	}
	if err := os.WriteFile(outputPath, append(payload, '\n'), 0o644); err != nil {
		return "", DatasetArtifact{}, fmt.Errorf("write dataset artifact: %w", err)
	}
	return outputPath, artifact, nil
}

func resolveTarget(spec Spec) (ast.Expr, DatasetTarget, error) {
	var expr ast.Expr
	var err error

	switch spec.Target.Kind {
	case TargetKindConcept:
		registry := concepts.StandardLibrary()
		expr, err = registry.ExpandSymbolic(spec.Target.Concept)
		if err != nil {
			return nil, DatasetTarget{}, fmt.Errorf("expand concept target %q: %w", spec.Target.Concept, err)
		}
	case TargetKindRaw:
		expr, err = parser.ParseString(spec.Target.RawEML)
		if err != nil {
			return nil, DatasetTarget{}, fmt.Errorf("parse raw target %q: %w", spec.Target.RawEML, err)
		}
	default:
		return nil, DatasetTarget{}, fmt.Errorf("unsupported target kind %q", spec.Target.Kind)
	}

	candidate := search.NewCandidate(expr)
	return candidate.Normalized, DatasetTarget{
		Kind:         spec.Target.Kind,
		Concept:      spec.Target.Concept,
		RawEML:       spec.Target.RawEML,
		CanonicalKey: candidate.Key,
		Expression:   candidate.Normalized.String(),
	}, nil
}

func samplePointsForDataset(dataset DatasetSpec) ([]float64, DatasetDomain, error) {
	switch dataset.Mode {
	case DatasetModeRealGrid:
		grid := dataset.Grid
		if grid == nil {
			return nil, DatasetDomain{}, fmt.Errorf("real_grid dataset requires grid")
		}
		points := realGridPoints(*grid)
		return points, DatasetDomain{Grid: grid}, nil
	case DatasetModeExplicitPoints:
		points := make([]float64, len(dataset.Points))
		copy(points, dataset.Points)
		return points, DatasetDomain{Points: points}, nil
	default:
		return nil, DatasetDomain{}, fmt.Errorf("unsupported dataset mode %q", dataset.Mode)
	}
}

func realGridPoints(grid RealGrid) []float64 {
	if grid.Count <= 0 {
		return nil
	}
	if grid.Count == 1 {
		return []float64{grid.Start}
	}
	step := (grid.Stop - grid.Start) / float64(grid.Count-1)
	points := make([]float64, 0, grid.Count)
	for i := 0; i < grid.Count; i++ {
		points = append(points, grid.Start+float64(i)*step)
	}
	return points
}
