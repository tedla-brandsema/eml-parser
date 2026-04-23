package experiment

import (
	"strings"
	"testing"
)

func TestParseSpecConceptTargetWithGrid(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "exp_real_grid",
		"description": "Recover exp(x) on a small real grid",
		"target": {
			"kind": "concept",
			"concept": "exp"
		},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {
				"start": -1,
				"stop": 1,
				"count": 5
			}
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 2,
				"max_nodes": 3
			},
			"top_n": 5
		},
		"recovery": {
			"expected_class": "exact_normalized_recovery",
			"expected_canonical_key": "eml(x, 1)"
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}
	if spec.ID != "exp_real_grid" {
		t.Fatalf("unexpected spec id: %+v", spec)
	}
	if spec.Target.Kind != TargetKindConcept || spec.Target.Concept != "exp" {
		t.Fatalf("unexpected target: %+v", spec.Target)
	}
	if spec.Dataset.Mode != DatasetModeRealGrid || spec.Dataset.Grid == nil {
		t.Fatalf("unexpected dataset: %+v", spec.Dataset)
	}
	if spec.Search.Mode != SearchModeEnumerativeReal {
		t.Fatalf("unexpected search mode: %+v", spec.Search)
	}
	if spec.Recovery.ExpectedClass != RecoveryClassExactNormalized {
		t.Fatalf("unexpected recovery: %+v", spec.Recovery)
	}
}

func TestParseSpecRawTargetWithPoints(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "raw_identity_points",
		"description": "Recover x from explicit points",
		"target": {
			"kind": "raw",
			"raw_eml": "x"
		},
		"dataset": {
			"mode": "real_points",
			"variable": "x",
			"points": [-1, 0, 1]
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 1,
				"max_nodes": 1
			},
			"top_n": 3
		},
		"recovery": {
			"expected_class": "no_recovery"
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}
	if spec.Target.Kind != TargetKindRaw || spec.Target.RawEML != "x" {
		t.Fatalf("unexpected raw target: %+v", spec.Target)
	}
	if spec.Dataset.Mode != DatasetModeExplicitPoints || len(spec.Dataset.Points) != 3 {
		t.Fatalf("unexpected dataset: %+v", spec.Dataset)
	}
}

func TestParseSpecRejectsInvalidSearchMode(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "bad_search_mode",
		"description": "invalid",
		"target": {
			"kind": "concept",
			"concept": "exp"
		},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {
				"start": -1,
				"stop": 1,
				"count": 5
			}
		},
		"search": {
			"mode": "beam_search",
			"bounds": {
				"max_depth": 2,
				"max_nodes": 3
			},
			"top_n": 5
		},
		"recovery": {
			"expected_class": "exact_normalized_recovery",
			"expected_canonical_key": "eml(x, 1)"
		}
	}`))
	if err == nil {
		t.Fatal("expected error for invalid search mode")
	}
	if !strings.Contains(err.Error(), SearchModeEnumerativeReal) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseSpecRejectsConceptEquivalentWithoutAllowedKeys(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "missing_equivalents",
		"description": "invalid",
		"target": {
			"kind": "concept",
			"concept": "exp"
		},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {
				"start": -1,
				"stop": 1,
				"count": 5
			}
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 2,
				"max_nodes": 3
			},
			"top_n": 5
		},
		"recovery": {
			"expected_class": "concept_equivalent_recovery",
			"expected_canonical_key": "eml(x, 1)"
		}
	}`))
	if err == nil {
		t.Fatal("expected error for missing allowed equivalent keys")
	}
	if !strings.Contains(err.Error(), "allowed_equivalent_keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseSpecRejectsApproximateOnlyWithoutThreshold(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "missing_threshold",
		"description": "invalid",
		"target": {
			"kind": "raw",
			"raw_eml": "x"
		},
		"dataset": {
			"mode": "real_points",
			"variable": "x",
			"points": [0, 1]
		},
		"search": {
			"mode": "enumerative_real",
			"bounds": {
				"max_depth": 1,
				"max_nodes": 1
			},
			"top_n": 2
		},
		"recovery": {
			"expected_class": "approximate_only_recovery"
		}
	}`))
	if err == nil {
		t.Fatal("expected error for missing approximate threshold")
	}
	if !strings.Contains(err.Error(), "approximate_threshold") {
		t.Fatalf("unexpected error: %v", err)
	}
}
