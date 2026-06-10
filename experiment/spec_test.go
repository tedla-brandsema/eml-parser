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

func TestParseSpecMazeRealMode(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "maze_smoke",
		"description": "Maze mode parses with snippet anchors and coverage",
		"target": {
			"kind": "raw",
			"raw_eml": "eml(eml(x, 1), 1)"
		},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {
				"start": 0.05,
				"stop": 0.35,
				"count": 8
			}
		},
		"search": {
			"mode": "maze_real",
			"bounds": {
				"max_depth": 3,
				"max_nodes": 5
			},
			"top_n": 5,
			"maze": {
				"snippet_artifact": "artifacts/snippets/raw_exp3.snippet.json",
				"snippet_ids": ["raw_exp3_exp1"],
				"accept_threshold": 1.0,
				"retain_threshold": 2.0,
				"coverage": {
					"min_window_size": 4,
					"coverage_weight": 0.25
				}
			}
		},
		"recovery": {
			"expected_class": "partial_coverage_recovery",
			"min_coverage_ratio": 0.5,
			"max_local_error": 0.01
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}
	if spec.Search.Mode != SearchModeMazeReal || spec.Search.Maze == nil {
		t.Fatalf("unexpected search: %+v", spec.Search)
	}
	if spec.Search.Maze.Coverage == nil || spec.Search.Maze.Coverage.MinWindowSize != 4 {
		t.Fatalf("unexpected coverage: %+v", spec.Search.Maze.Coverage)
	}
	if spec.Recovery.ExpectedClass != RecoveryClassPartialCoverage {
		t.Fatalf("unexpected recovery: %+v", spec.Recovery)
	}
}

func TestParseSpecMazeModeRequiresMazeBlock(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "maze_missing_block",
		"description": "Maze mode without maze block must fail",
		"target": {"kind": "raw", "raw_eml": "x"},
		"dataset": {"mode": "real_points", "variable": "x", "points": [0.1, 0.2]},
		"search": {
			"mode": "maze_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 3
		},
		"recovery": {"expected_class": "no_recovery"}
	}`))
	if err == nil || !strings.Contains(err.Error(), "requires maze block") {
		t.Fatalf("expected maze block error, got %v", err)
	}
}

func TestParseSpecEnumerativeModeRejectsMazeBlock(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "enumerative_with_maze",
		"description": "Enumerative mode with maze block must fail",
		"target": {"kind": "raw", "raw_eml": "x"},
		"dataset": {"mode": "real_points", "variable": "x", "points": [0.1, 0.2]},
		"search": {
			"mode": "enumerative_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 3,
			"maze": {
				"snippet_artifact": "artifacts/snippets/raw_exp3.snippet.json",
				"accept_threshold": 1.0,
				"retain_threshold": 2.0
			}
		},
		"recovery": {"expected_class": "no_recovery"}
	}`))
	if err == nil || !strings.Contains(err.Error(), "must not set maze") {
		t.Fatalf("expected maze rejection error, got %v", err)
	}
}

func TestParseSpecMazeRecoveryClassRequiresMazeMode(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "snippet_class_wrong_mode",
		"description": "Snippet recovery class with enumerative mode must fail",
		"target": {"kind": "raw", "raw_eml": "x"},
		"dataset": {"mode": "real_points", "variable": "x", "points": [0.1, 0.2]},
		"search": {
			"mode": "enumerative_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 3
		},
		"recovery": {
			"expected_class": "snippet_recovery",
			"expected_snippet_keys": ["eml(x, 1)"]
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "requires search mode") {
		t.Fatalf("expected search mode error, got %v", err)
	}
}

func TestParseSpecPartialCoverageRequiresCoverageScoring(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "partial_coverage_no_scoring",
		"description": "Partial coverage class without coverage scoring must fail",
		"target": {"kind": "raw", "raw_eml": "x"},
		"dataset": {"mode": "real_points", "variable": "x", "points": [0.1, 0.2]},
		"search": {
			"mode": "maze_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 3,
			"maze": {
				"snippet_artifact": "artifacts/snippets/raw_exp3.snippet.json",
				"accept_threshold": 1.0,
				"retain_threshold": 2.0
			}
		},
		"recovery": {
			"expected_class": "partial_coverage_recovery",
			"min_coverage_ratio": 0.5,
			"max_local_error": 0.01
		}
	}`))
	if err == nil || !strings.Contains(err.Error(), "requires maze coverage scoring") {
		t.Fatalf("expected coverage scoring error, got %v", err)
	}
}

func TestParseSpecWindowSetCoverage(t *testing.T) {
	spec, err := ParseSpec([]byte(`{
		"id": "window_set_smoke",
		"description": "Window-set coverage parses with tolerance and count bound",
		"target": {"kind": "concept", "concept": "sin"},
		"dataset": {
			"mode": "real_grid",
			"variable": "x",
			"grid": {"start": 1.2, "stop": 8.2, "count": 36}
		},
		"search": {
			"mode": "maze_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 5,
			"maze": {
				"snippet_artifact": "artifacts/snippets/raw_exp3.snippet.json",
				"accept_threshold": 1.0,
				"retain_threshold": 2.0,
				"coverage": {
					"mode": "window_set",
					"min_window_size": 3,
					"point_tolerance": 0.02,
					"max_window_count": 2
				}
			}
		},
		"recovery": {
			"expected_class": "partial_coverage_recovery",
			"min_coverage_ratio": 0.25,
			"max_local_error": 0.02
		}
	}`))
	if err != nil {
		t.Fatalf("ParseSpec returned error: %v", err)
	}
	if spec.Search.Maze.Coverage.Mode != CoverageModeWindowSet {
		t.Fatalf("unexpected coverage mode: %+v", spec.Search.Maze.Coverage)
	}
}

func TestParseSpecWindowSetCoverageRequiresTolerance(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "window_set_missing_tolerance",
		"description": "Window-set coverage without point tolerance must fail",
		"target": {"kind": "raw", "raw_eml": "x"},
		"dataset": {"mode": "real_points", "variable": "x", "points": [0.1, 0.2, 0.3]},
		"search": {
			"mode": "maze_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 3,
			"maze": {
				"snippet_artifact": "artifacts/snippets/raw_exp3.snippet.json",
				"accept_threshold": 1.0,
				"retain_threshold": 2.0,
				"coverage": {
					"mode": "window_set",
					"min_window_size": 3,
					"max_window_count": 2
				}
			}
		},
		"recovery": {"expected_class": "no_recovery"}
	}`))
	if err == nil || !strings.Contains(err.Error(), "requires positive point_tolerance") {
		t.Fatalf("expected point_tolerance error, got %v", err)
	}
}

func TestParseSpecSingleWindowCoverageRejectsWindowSetFields(t *testing.T) {
	_, err := ParseSpec([]byte(`{
		"id": "single_window_with_tolerance",
		"description": "Single-window coverage with window-set fields must fail",
		"target": {"kind": "raw", "raw_eml": "x"},
		"dataset": {"mode": "real_points", "variable": "x", "points": [0.1, 0.2, 0.3]},
		"search": {
			"mode": "maze_real",
			"bounds": {"max_depth": 2, "max_nodes": 3},
			"top_n": 3,
			"maze": {
				"snippet_artifact": "artifacts/snippets/raw_exp3.snippet.json",
				"accept_threshold": 1.0,
				"retain_threshold": 2.0,
				"coverage": {
					"min_window_size": 3,
					"point_tolerance": 0.02
				}
			}
		},
		"recovery": {"expected_class": "no_recovery"}
	}`))
	if err == nil || !strings.Contains(err.Error(), "must not set point_tolerance") {
		t.Fatalf("expected single-window rejection error, got %v", err)
	}
}
