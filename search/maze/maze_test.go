package maze

import (
	"testing"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/eval"
	"eml-parser/search/common"
)

func TestMazeRealSearchGrowsTowardExp(t *testing.T) {
	fixtures, err := common.RealBenchmarkFixtures()
	if err != nil {
		t.Fatalf("RealBenchmarkFixtures returned error: %v", err)
	}
	fixture := fixtures[0]

	report, err := MazeRealSearch(fixture, eval.Complex128Backend{}, []Anchor{
		{Name: "x_anchor", Expr: ast.Variable{Name: "x"}},
	}, MazeOptions{
		Bounds:          common.Bounds{MaxDepth: 2, MaxNodes: 3},
		TopN:            3,
		AcceptThreshold: 0.1,
		RetainThreshold: 2.0,
	})
	if err != nil {
		t.Fatalf("MazeRealSearch returned error: %v", err)
	}
	if len(report.BestCandidates) == 0 {
		t.Fatal("expected best candidates")
	}
	if report.BestCandidates[0].Candidate.Key != "eml(x, 1)" {
		t.Fatalf("expected exp(x) best candidate, got %q", report.BestCandidates[0].Candidate.Key)
	}
}

func TestMazeRealSearchRetainsPartialResults(t *testing.T) {
	fixtures, err := common.RealBenchmarkFixtures()
	if err != nil {
		t.Fatalf("RealBenchmarkFixtures returned error: %v", err)
	}
	fixture := fixtures[0]

	report, err := MazeRealSearch(fixture, eval.Complex128Backend{}, []Anchor{
		{Name: "one_anchor", Expr: ast.One{}},
	}, MazeOptions{
		Bounds:          common.Bounds{MaxDepth: 2, MaxNodes: 3},
		TopN:            3,
		AcceptThreshold: 0.01,
		RetainThreshold: 5.0,
		Atoms:           []ast.Expr{ast.One{}},
	})
	if err != nil {
		t.Fatalf("MazeRealSearch returned error: %v", err)
	}
	if len(report.PartialResults) == 0 {
		t.Fatal("expected retained partial results")
	}
}

func TestMazeRealSearchDeterministic(t *testing.T) {
	fixtures, err := common.RealBenchmarkFixtures()
	if err != nil {
		t.Fatalf("RealBenchmarkFixtures returned error: %v", err)
	}
	fixture := fixtures[0]
	anchors := []Anchor{{Name: "x_anchor", Expr: ast.Variable{Name: "x"}}}
	options := MazeOptions{
		Bounds:          common.Bounds{MaxDepth: 2, MaxNodes: 3},
		TopN:            3,
		AcceptThreshold: 0.1,
		RetainThreshold: 2.0,
	}

	first, err := MazeRealSearch(fixture, eval.Complex128Backend{}, anchors, options)
	if err != nil {
		t.Fatalf("first MazeRealSearch error: %v", err)
	}
	second, err := MazeRealSearch(fixture, eval.Complex128Backend{}, anchors, options)
	if err != nil {
		t.Fatalf("second MazeRealSearch error: %v", err)
	}
	if len(first.BestCandidates) != len(second.BestCandidates) {
		t.Fatalf("best candidate counts differ: %d vs %d", len(first.BestCandidates), len(second.BestCandidates))
	}
	for i := range first.BestCandidates {
		if first.BestCandidates[i].Candidate.Key != second.BestCandidates[i].Candidate.Key || first.BestCandidates[i].Score != second.BestCandidates[i].Score {
			t.Fatalf("deterministic outputs differ at %d", i)
		}
	}
}

func TestMazeRealSearchMultiAnchorSurvival(t *testing.T) {
	registry := concepts.StandardLibrary()
	expExpr, err := registry.ExpandSymbolic("exp")
	if err != nil {
		t.Fatalf("ExpandSymbolic returned error: %v", err)
	}
	fixtures, err := common.RealBenchmarkFixtures()
	if err != nil {
		t.Fatalf("RealBenchmarkFixtures returned error: %v", err)
	}
	fixture := fixtures[0]

	report, err := MazeRealSearch(fixture, eval.Complex128Backend{}, []Anchor{
		{Name: "x_anchor", Expr: ast.Variable{Name: "x"}},
		{Name: "exp_anchor", Expr: expExpr},
	}, MazeOptions{
		Bounds:          common.Bounds{MaxDepth: 2, MaxNodes: 3},
		TopN:            5,
		AcceptThreshold: 0.5,
		RetainThreshold: 2.0,
	})
	if err != nil {
		t.Fatalf("MazeRealSearch returned error: %v", err)
	}
	if report.Diagnostics.AnchorCount != 2 {
		t.Fatalf("unexpected anchor count: %d", report.Diagnostics.AnchorCount)
	}
	if report.Diagnostics.ThreadsSpawned < 2 {
		t.Fatalf("expected at least two spawned threads, got %d", report.Diagnostics.ThreadsSpawned)
	}
}
