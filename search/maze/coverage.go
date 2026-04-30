package maze

import (
	"eml-parser/eval"
	"eml-parser/family"
	"eml-parser/search/common"
)

type CoverageOptions = common.PartialCoverageOptions

func MazeRealSearchPartialCoverage(fixture common.BenchmarkCase[float64], backend eval.Backend[complex128], anchors []Anchor, options MazeOptions, coverage CoverageOptions) (MazeReport, error) {
	if options.TopN <= 0 {
		options.TopN = 5
	}
	if options.AcceptThreshold == 0 {
		options.AcceptThreshold = 0.5
	}
	if options.RetainThreshold == 0 {
		options.RetainThreshold = 2.0
	}
	if options.MinImprovement == 0 {
		options.MinImprovement = 1e-9
	}
	if coverage.MinWindowSize <= 0 {
		coverage.MinWindowSize = 3
	}

	target := common.NewSearchTarget([]string{fixture.TargetKey}, fixture.Samples)
	scorer := common.RealPartialCoverageScorer{Options: coverage}
	retention := common.CoverageRetentionPolicy{
		AcceptThreshold: options.AcceptThreshold,
		RetainThreshold: options.RetainThreshold,
		MinImprovement:  options.MinImprovement,
		MinCoveredCount: coverage.MinWindowSize,
	}
	return MazeRealSearchWithPolicies(target, backend, anchors, options, scorer, retention)
}

func MazeRealSearchFromSnippetArtifactPartialCoverage(artifact family.SnippetDatasetArtifact, backend eval.Backend[complex128], snippetIDs []string, options MazeOptions, coverage CoverageOptions) (MazeReport, error) {
	anchors, err := AnchorsFromSnippetArtifact(artifact, snippetIDs...)
	if err != nil {
		return MazeReport{}, err
	}
	fixture, err := benchmarkCaseFromSnippetArtifact(artifact)
	if err != nil {
		return MazeReport{}, err
	}
	return MazeRealSearchPartialCoverage(fixture, backend, anchors, options, coverage)
}
