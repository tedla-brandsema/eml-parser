package experiment

import (
	"testing"

	"eml-parser/ast"
	"eml-parser/search"
	"eml-parser/search/maze"
)

func TestClassifyRecoveryExactNormalized(t *testing.T) {
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:        RecoveryClassExactNormalized,
			ExpectedCanonicalKey: "x",
		},
	}
	report := search.SearchReport{
		Results: []search.SearchResult{
			{
				Candidate: search.NewCandidate(ast.Variable{Name: "x"}),
				Score:     0,
			},
		},
	}
	if got := ClassifyRecovery(spec, report); got != RecoveryClassExactNormalized {
		t.Fatalf("expected %q, got %q", RecoveryClassExactNormalized, got)
	}
}

func TestClassifyRecoveryConceptEquivalent(t *testing.T) {
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:         RecoveryClassConceptEquivalent,
			ExpectedCanonicalKey:  "eml(x, 1)",
			AllowedEquivalentKeys: []string{"x"},
		},
	}
	report := search.SearchReport{
		Results: []search.SearchResult{
			{
				Candidate: search.NewCandidate(ast.Variable{Name: "x"}),
				Score:     0,
			},
		},
	}
	if got := ClassifyRecovery(spec, report); got != RecoveryClassConceptEquivalent {
		t.Fatalf("expected %q, got %q", RecoveryClassConceptEquivalent, got)
	}
}

func TestClassifyRecoveryApproximateOnly(t *testing.T) {
	threshold := 0.01
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:        RecoveryClassApproximateOnly,
			ApproximateThreshold: &threshold,
		},
	}
	report := search.SearchReport{
		Results: []search.SearchResult{
			{
				Candidate: search.NewCandidate(ast.One{}),
				Score:     0.001,
			},
		},
	}
	if got := ClassifyRecovery(spec, report); got != RecoveryClassApproximateOnly {
		t.Fatalf("expected %q, got %q", RecoveryClassApproximateOnly, got)
	}
}

func TestClassifyRecoveryNoRecovery(t *testing.T) {
	threshold := 0.001
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:        RecoveryClassApproximateOnly,
			ApproximateThreshold: &threshold,
		},
	}
	report := search.SearchReport{
		Results: []search.SearchResult{
			{
				Candidate: search.NewCandidate(ast.One{}),
				Score:     1,
			},
		},
	}
	if got := ClassifyRecovery(spec, report); got != RecoveryClassNoRecovery {
		t.Fatalf("expected %q, got %q", RecoveryClassNoRecovery, got)
	}
}

func mazeReportWithCandidates(candidates ...maze.CandidateScore) maze.MazeReport {
	return maze.MazeReport{BestCandidates: candidates}
}

func TestClassifyMazeRecoveryFullLaw(t *testing.T) {
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:        RecoveryClassFullLaw,
			ExpectedCanonicalKey: "eml(x, 1)",
		},
	}
	report := mazeReportWithCandidates(maze.CandidateScore{
		Candidate: search.NewCandidate(ast.Apply{Left: ast.Variable{Name: "x"}, Right: ast.One{}}),
		Score:     0,
	})
	if got := ClassifyMazeRecovery(spec, report); got != RecoveryClassFullLaw {
		t.Fatalf("expected %q, got %q", RecoveryClassFullLaw, got)
	}
}

func TestClassifyMazeRecoverySnippetInTopN(t *testing.T) {
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:       RecoveryClassSnippet,
			ExpectedSnippetKeys: []string{"eml(x, 1)"},
		},
	}
	report := mazeReportWithCandidates(
		maze.CandidateScore{
			Candidate: search.NewCandidate(ast.Variable{Name: "x"}),
			Score:     1.0,
		},
		maze.CandidateScore{
			Candidate: search.NewCandidate(ast.Apply{Left: ast.Variable{Name: "x"}, Right: ast.One{}}),
			Score:     2.0,
		},
	)
	if got := ClassifyMazeRecovery(spec, report); got != RecoveryClassSnippet {
		t.Fatalf("expected %q, got %q", RecoveryClassSnippet, got)
	}
}

func TestClassifyMazeRecoveryPartialCoverageThresholds(t *testing.T) {
	minCoverage := 0.4
	maxLocalError := 0.05
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:    RecoveryClassPartialCoverage,
			MinCoverageRatio: &minCoverage,
			MaxLocalError:    &maxLocalError,
		},
	}
	report := mazeReportWithCandidates(maze.CandidateScore{
		Candidate: search.NewCandidate(ast.Variable{Name: "x"}),
		Score:     0.2,
		ScoreDetails: search.ScoreResult{
			Finite:        true,
			CoverageRatio: 0.5,
			LocalError:    0.01,
		},
	})
	if got := ClassifyMazeRecovery(spec, report); got != RecoveryClassPartialCoverage {
		t.Fatalf("expected %q, got %q", RecoveryClassPartialCoverage, got)
	}

	failing := mazeReportWithCandidates(maze.CandidateScore{
		Candidate: search.NewCandidate(ast.Variable{Name: "x"}),
		Score:     0.2,
		ScoreDetails: search.ScoreResult{
			Finite:        true,
			CoverageRatio: 0.3,
			LocalError:    0.01,
		},
	})
	if got := ClassifyMazeRecovery(spec, failing); got != RecoveryClassNoRecovery {
		t.Fatalf("expected %q, got %q", RecoveryClassNoRecovery, got)
	}
}

func TestClassifyMazeRecoveryFullLawOutranksSnippet(t *testing.T) {
	spec := Spec{
		Recovery: RecoverySpec{
			ExpectedClass:        RecoveryClassFullLaw,
			ExpectedCanonicalKey: "eml(x, 1)",
			ExpectedSnippetKeys:  []string{"eml(x, 1)"},
		},
	}
	report := mazeReportWithCandidates(maze.CandidateScore{
		Candidate: search.NewCandidate(ast.Apply{Left: ast.Variable{Name: "x"}, Right: ast.One{}}),
		Score:     0,
	})
	if got := ClassifyMazeRecovery(spec, report); got != RecoveryClassFullLaw {
		t.Fatalf("expected %q, got %q", RecoveryClassFullLaw, got)
	}
}

func TestClassifyMazeRecoveryEmptyReport(t *testing.T) {
	spec := Spec{Recovery: RecoverySpec{ExpectedClass: RecoveryClassNoRecovery}}
	if got := ClassifyMazeRecovery(spec, maze.MazeReport{}); got != RecoveryClassNoRecovery {
		t.Fatalf("expected %q, got %q", RecoveryClassNoRecovery, got)
	}
}
