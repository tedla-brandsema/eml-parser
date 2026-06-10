package experiment

import (
	"eml-parser/search"
	"eml-parser/search/maze"
)

// ClassifyRecovery assigns one recovery class to a completed search report
// using the experiment spec's declared recovery criteria.
func ClassifyRecovery(spec Spec, report search.SearchReport) string {
	if len(report.Results) == 0 {
		return RecoveryClassNoRecovery
	}

	top := report.Results[0]
	if spec.Recovery.ExpectedClass == RecoveryClassExactNormalized || spec.Recovery.ExpectedCanonicalKey != "" {
		if top.Candidate.Key == spec.Recovery.ExpectedCanonicalKey {
			return RecoveryClassExactNormalized
		}
	}

	if len(spec.Recovery.AllowedEquivalentKeys) > 0 {
		for _, key := range spec.Recovery.AllowedEquivalentKeys {
			if top.Candidate.Key == key {
				return RecoveryClassConceptEquivalent
			}
		}
	}

	if spec.Recovery.ApproximateThreshold != nil && top.Score <= *spec.Recovery.ApproximateThreshold {
		return RecoveryClassApproximateOnly
	}

	return RecoveryClassNoRecovery
}

// ClassifyMazeRecovery assigns one partial recovery class to a completed maze
// report. Classification order: full law first, then declared snippet keys,
// then partial coverage, then no recovery. Full-law and partial-coverage
// checks use only the top-ranked candidate; snippet recovery accepts a
// declared snippet key anywhere in the returned top-N because fractional
// recovery is about which labeled pieces survive, not only which ranks first.
func ClassifyMazeRecovery(spec Spec, report maze.MazeReport) string {
	if len(report.BestCandidates) == 0 {
		return RecoveryClassNoRecovery
	}

	top := report.BestCandidates[0]
	if spec.Recovery.ExpectedCanonicalKey != "" && top.Candidate.Key == spec.Recovery.ExpectedCanonicalKey {
		return RecoveryClassFullLaw
	}

	if len(spec.Recovery.ExpectedSnippetKeys) > 0 {
		for _, candidate := range report.BestCandidates {
			for _, key := range spec.Recovery.ExpectedSnippetKeys {
				if candidate.Candidate.Key == key {
					return RecoveryClassSnippet
				}
			}
		}
	}

	if spec.Recovery.MinCoverageRatio != nil && spec.Recovery.MaxLocalError != nil {
		details := top.ScoreDetails
		if details.Finite &&
			details.CoverageRatio >= *spec.Recovery.MinCoverageRatio &&
			details.LocalError <= *spec.Recovery.MaxLocalError {
			return RecoveryClassPartialCoverage
		}
	}

	return RecoveryClassNoRecovery
}
