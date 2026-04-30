package maze

import (
	"testing"

	"eml-parser/ast"
	"eml-parser/concepts"
	"eml-parser/eval"
	"eml-parser/family"
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

func TestMazeRealSearchExpandsNonRootFrontier(t *testing.T) {
	expr := ast.Apply{
		Left: ast.Apply{
			Left:  ast.Variable{Name: "x"},
			Right: ast.One{},
		},
		Right: ast.One{},
	}
	thread := GrowthThread{
		AnchorName:  "nested_anchor",
		Current:     common.NewCandidate(expr),
		Score:       0.0,
		ParentScore: 0.0,
		Frontiers:   openFrontiers(expr),
		Status:      ThreadStatusActive,
	}

	children, partials, _, err := expandThread(
		thread,
		[]ast.Expr{ast.Variable{Name: "x"}},
		common.NewSearchTarget([]string{"x"}, []common.Sample[float64]{
			{Vars: map[string]float64{"x": 0}, Target: 1},
		}),
		eval.Complex128Backend{},
		MazeOptions{
			Bounds:          common.Bounds{MaxDepth: 3, MaxNodes: 7},
			AcceptThreshold: 1e9,
			RetainThreshold: 1e9,
			MinImprovement:  -10.0,
		},
		common.RealMSEScorer{},
		common.ThresholdRetentionPolicy{
			AcceptThreshold: 1e9,
			RetainThreshold: 1e9,
			MinImprovement:  -10.0,
		},
		map[string]bool{thread.Current.Key: true},
		&MazeDiagnostics{},
	)
	if err != nil {
		t.Fatalf("expandThread returned error: %v", err)
	}
	if len(children) == 0 {
		t.Fatal("expected frontier expansions")
	}
	foundNonRoot := false
	for _, child := range children {
		if len(child.History) == 0 {
			continue
		}
		if child.History[len(child.History)-1].FrontierPath != "root" {
			foundNonRoot = true
			break
		}
	}
	for _, partial := range partials {
		if len(partial.History) == 0 {
			continue
		}
		if partial.History[len(partial.History)-1].FrontierPath != "root" {
			foundNonRoot = true
			break
		}
	}
	if !foundNonRoot {
		t.Fatal("expected at least one expansion from a non-root frontier")
	}
}

func TestMazeRealSearchTracksMultipleOpenFrontiers(t *testing.T) {
	expr := ast.Apply{Left: ast.Variable{Name: "x"}, Right: ast.One{}}
	frontiers := openFrontiers(expr)
	if len(frontiers) < 3 {
		t.Fatalf("expected multiple frontiers, got %d", len(frontiers))
	}
	if frontiers[0].Path != "root" || frontiers[1].Path != "root.L" || frontiers[2].Path != "root.R" {
		t.Fatalf("unexpected frontier ordering: %#v", frontiers[:3])
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

func TestMazeRealSearchStalledBranchBecomesPartial(t *testing.T) {
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
		AcceptThreshold: 10.0,
		RetainThreshold: 10.0,
		MinImprovement:  100.0,
		Atoms:           []ast.Expr{ast.One{}},
	})
	if err != nil {
		t.Fatalf("MazeRealSearch returned error: %v", err)
	}
	if len(report.PartialResults) == 0 {
		t.Fatal("expected partial results")
	}
	foundStalled := false
	for _, partial := range report.PartialResults {
		if partial.Reason == "stalled" {
			foundStalled = true
			break
		}
	}
	if !foundStalled {
		t.Fatal("expected a stalled partial result")
	}
}

func TestAnchorsFromSnippetArtifactPreservesProvenance(t *testing.T) {
	artifact, err := family.GenerateSnippetDataset(
		family.CuratedSnippetTargets()[0],
		concepts.StandardLibrary(),
		[]family.SamplingDomain{
			{
				DomainID: "default",
				Sampling: family.SamplingSpec{
					Variable:    "x",
					Start:       0.05,
					Stop:        0.25,
					PointCount:  4,
					SampleCount: 1,
					Seed:        0,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateSnippetDataset returned error: %v", err)
	}

	anchors, err := AnchorsFromSnippetArtifact(artifact, artifact.Snippets[0].SnippetID)
	if err != nil {
		t.Fatalf("AnchorsFromSnippetArtifact returned error: %v", err)
	}
	if len(anchors) != 1 {
		t.Fatalf("expected one anchor, got %d", len(anchors))
	}
	if anchors[0].Provenance == nil {
		t.Fatal("expected snippet provenance")
	}
	if anchors[0].Provenance.SourceKind != AnchorSourceSnippet {
		t.Fatalf("unexpected source kind: %q", anchors[0].Provenance.SourceKind)
	}
	if anchors[0].Provenance.SnippetID != artifact.Snippets[0].SnippetID {
		t.Fatalf("unexpected snippet id: %q", anchors[0].Provenance.SnippetID)
	}
}

func TestMazeRealSearchFromSnippetArtifact(t *testing.T) {
	artifact, err := family.GenerateSnippetDataset(
		family.CuratedSnippetTargets()[0],
		concepts.StandardLibrary(),
		[]family.SamplingDomain{
			{
				DomainID: "default",
				Sampling: family.SamplingSpec{
					Variable:    "x",
					Start:       0.05,
					Stop:        0.25,
					PointCount:  4,
					SampleCount: 1,
					Seed:        0,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateSnippetDataset returned error: %v", err)
	}

	report, err := MazeRealSearchFromSnippetArtifact(
		artifact,
		eval.Complex128Backend{},
		[]string{artifact.Snippets[0].SnippetID},
		MazeOptions{
			Bounds:          common.Bounds{MaxDepth: 5, MaxNodes: 9},
			TopN:            5,
			AcceptThreshold: 10.0,
			RetainThreshold: 10.0,
		},
	)
	if err != nil {
		t.Fatalf("MazeRealSearchFromSnippetArtifact returned error: %v", err)
	}
	if len(report.BestCandidates) == 0 {
		t.Fatal("expected snippet-seeded maze candidates")
	}
	if report.BestCandidates[0].Provenance == nil {
		t.Fatal("expected provenance on best candidate")
	}
	if report.BestCandidates[0].Provenance.SourceKind != AnchorSourceSnippet {
		t.Fatalf("unexpected source kind: %q", report.BestCandidates[0].Provenance.SourceKind)
	}
}

func TestMatchSnippetAnchorsRanksExactWholeTraceFirst(t *testing.T) {
	artifact, err := family.GenerateSnippetDataset(
		family.CuratedSnippetTargets()[0],
		concepts.StandardLibrary(),
		[]family.SamplingDomain{
			{
				DomainID: "default",
				Sampling: family.SamplingSpec{
					Variable:    "x",
					Start:       0.05,
					Stop:        0.25,
					PointCount:  4,
					SampleCount: 1,
					Seed:        0,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateSnippetDataset returned error: %v", err)
	}

	target, err := SearchTargetFromSnippetTrace(artifact, artifact.Snippets[0].SnippetID, "default", "")
	if err != nil {
		t.Fatalf("SearchTargetFromSnippetTrace returned error: %v", err)
	}

	report, err := MatchSnippetAnchors(target, []family.SnippetDatasetArtifact{artifact}, SpawnOptions{
		TopK:     2,
		MaxScore: 1e-12,
	})
	if err != nil {
		t.Fatalf("MatchSnippetAnchors returned error: %v", err)
	}
	if len(report.Matches) == 0 {
		t.Fatal("expected ranked snippet matches")
	}
	if report.Matches[0].SnippetID != artifact.Snippets[0].SnippetID {
		t.Fatalf("expected exact snippet first, got %q", report.Matches[0].SnippetID)
	}
	if report.Matches[0].Score > 1e-12 {
		t.Fatalf("expected near-zero exact match score, got %g", report.Matches[0].Score)
	}
	if len(report.Anchors) == 0 {
		t.Fatal("expected promoted anchors")
	}
}

func TestMatchSnippetAnchorsThresholdRejects(t *testing.T) {
	artifact, err := family.GenerateSnippetDataset(
		family.CuratedSnippetTargets()[0],
		concepts.StandardLibrary(),
		[]family.SamplingDomain{
			{
				DomainID: "default",
				Sampling: family.SamplingSpec{
					Variable:    "x",
					Start:       0.05,
					Stop:        0.25,
					PointCount:  4,
					SampleCount: 1,
					Seed:        0,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateSnippetDataset returned error: %v", err)
	}

	target := common.NewSearchTarget([]string{"x"}, []common.Sample[float64]{
		{Vars: map[string]float64{"x": 0.1}, Target: 999},
		{Vars: map[string]float64{"x": 0.2}, Target: 999},
		{Vars: map[string]float64{"x": 0.3}, Target: 999},
		{Vars: map[string]float64{"x": 0.4}, Target: 999},
	})
	report, err := MatchSnippetAnchors(target, []family.SnippetDatasetArtifact{artifact}, SpawnOptions{
		TopK:     2,
		MaxScore: 1e-12,
	})
	if err != nil {
		t.Fatalf("MatchSnippetAnchors returned error: %v", err)
	}
	if len(report.Anchors) != 0 {
		t.Fatalf("expected no promoted anchors, got %d", len(report.Anchors))
	}
}

func TestMazeRealSearchFromSpawnedSnippets(t *testing.T) {
	artifact, err := family.GenerateSnippetDataset(
		family.CuratedSnippetTargets()[0],
		concepts.StandardLibrary(),
		[]family.SamplingDomain{
			{
				DomainID: "default",
				Sampling: family.SamplingSpec{
					Variable:    "x",
					Start:       0.05,
					Stop:        0.25,
					PointCount:  4,
					SampleCount: 1,
					Seed:        0,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("GenerateSnippetDataset returned error: %v", err)
	}

	target, err := SearchTargetFromSnippetTrace(artifact, artifact.Snippets[0].SnippetID, "default", "")
	if err != nil {
		t.Fatalf("SearchTargetFromSnippetTrace returned error: %v", err)
	}

	mazeReport, spawnReport, err := MazeRealSearchFromSpawnedSnippets(
		target,
		eval.Complex128Backend{},
		[]family.SnippetDatasetArtifact{artifact},
		SpawnOptions{
			TopK:     1,
			MaxScore: 1e-12,
		},
		MazeOptions{
			Bounds:          common.Bounds{MaxDepth: 5, MaxNodes: 9},
			TopN:            5,
			AcceptThreshold: 10.0,
			RetainThreshold: 10.0,
		},
	)
	if err != nil {
		t.Fatalf("MazeRealSearchFromSpawnedSnippets returned error: %v", err)
	}
	if len(spawnReport.Anchors) != 1 {
		t.Fatalf("expected one spawned anchor, got %d", len(spawnReport.Anchors))
	}
	if len(mazeReport.BestCandidates) == 0 {
		t.Fatal("expected maze candidates from spawned anchors")
	}
	if mazeReport.BestCandidates[0].Provenance == nil {
		t.Fatal("expected snippet provenance on maze result")
	}
}
