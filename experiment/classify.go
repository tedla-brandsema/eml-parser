package experiment

import "eml-parser/search"

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
