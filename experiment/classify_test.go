package experiment

import (
	"testing"

	"eml-parser/ast"
	"eml-parser/search"
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
