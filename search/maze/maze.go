package maze

import (
	"math"
	"sort"

	"eml-parser/ast"
	"eml-parser/eval"
	"eml-parser/search/common"
)

type Anchor struct {
	Name string
	Expr ast.Expr
}

type Frontier struct {
	Name string
}

type ExpansionStep struct {
	ParentKey string
	Direction string
	Atom      string
	Score     float64
	ResultKey string
	Frontier  string
	Improved  bool
}

type GrowthThread struct {
	AnchorName string
	Current    common.Candidate
	Score      float64
	Frontiers  []Frontier
	History    []ExpansionStep
	Active     bool
	Pruned     bool
	Completed  bool
}

type PartialResult struct {
	AnchorName string
	Candidate  common.Candidate
	Score      float64
	Reason     string
	History    []ExpansionStep
}

type MazeOptions struct {
	Bounds          common.Bounds
	TopN            int
	AcceptThreshold float64
	RetainThreshold float64
	Atoms           []ast.Expr
}

type MazeDiagnostics struct {
	AnchorCount           int
	ThreadsSpawned        int
	BranchesExpanded      int
	BranchesPruned        int
	BranchesRetained      int
	DuplicateEliminations int
	MaxDepthReached       int
	BestScore             float64
}

type CandidateScore struct {
	Candidate  common.Candidate
	Score      float64
	AnchorName string
	History    []ExpansionStep
}

type MazeReport struct {
	BestCandidates []CandidateScore
	PartialResults []PartialResult
	Diagnostics    MazeDiagnostics
}

func MazeRealSearch(fixture common.BenchmarkCase[float64], backend eval.Backend[complex128], anchors []Anchor, options MazeOptions) (MazeReport, error) {
	if len(anchors) == 0 {
		return MazeReport{}, nil
	}
	atoms := options.Atoms
	if len(atoms) == 0 {
		atoms = common.AtomicSeeds(fixture.TargetKey)
	}
	if options.TopN <= 0 {
		options.TopN = 5
	}
	if options.AcceptThreshold == 0 {
		options.AcceptThreshold = 0.5
	}
	if options.RetainThreshold == 0 {
		options.RetainThreshold = 2.0
	}

	report := MazeReport{
		Diagnostics: MazeDiagnostics{
			AnchorCount: len(anchors),
			BestScore:   math.Inf(1),
		},
	}

	var stack []GrowthThread
	seen := make(map[string]bool)

	for _, anchor := range anchors {
		candidate := common.NewCandidate(anchor.Expr)
		score, err := common.RealMSE(candidate, backend, fixture.Samples)
		if err != nil || !isFinite(score) {
			continue
		}
		thread := GrowthThread{
			AnchorName: anchor.Name,
			Current:    candidate,
			Score:      score,
			Frontiers:  []Frontier{{Name: "root"}},
			Active:     true,
		}
		stack = append(stack, thread)
		seen[candidate.Key] = true
		report.Diagnostics.ThreadsSpawned++
		if candidate.Stats.TreeDepth > report.Diagnostics.MaxDepthReached {
			report.Diagnostics.MaxDepthReached = candidate.Stats.TreeDepth
		}
		if score < report.Diagnostics.BestScore {
			report.Diagnostics.BestScore = score
		}
		report.BestCandidates = append(report.BestCandidates, CandidateScore{
			Candidate:  candidate,
			Score:      score,
			AnchorName: anchor.Name,
		})
	}

	for len(stack) > 0 {
		thread := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		children, partials, err := expandThread(thread, atoms, fixture.Samples, backend, options, seen, &report.Diagnostics)
		if err != nil {
			return MazeReport{}, err
		}
		report.PartialResults = append(report.PartialResults, partials...)
		report.Diagnostics.BranchesRetained += len(partials)
		if len(children) == 0 {
			if thread.Score <= options.RetainThreshold {
				report.PartialResults = append(report.PartialResults, PartialResult{
					AnchorName: thread.AnchorName,
					Candidate:  thread.Current,
					Score:      thread.Score,
					Reason:     "dead_end",
					History:    append([]ExpansionStep(nil), thread.History...),
				})
				report.Diagnostics.BranchesRetained++
			}
			continue
		}

		sort.Slice(children, func(i, j int) bool {
			if children[i].Score == children[j].Score {
				return children[i].Current.Key < children[j].Current.Key
			}
			return children[i].Score < children[j].Score
		})

		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, children[i])
			report.Diagnostics.ThreadsSpawned++
		}

		for _, child := range children {
			report.BestCandidates = append(report.BestCandidates, CandidateScore{
				Candidate:  child.Current,
				Score:      child.Score,
				AnchorName: child.AnchorName,
				History:    append([]ExpansionStep(nil), child.History...),
			})
			if child.Current.Stats.TreeDepth > report.Diagnostics.MaxDepthReached {
				report.Diagnostics.MaxDepthReached = child.Current.Stats.TreeDepth
			}
			if child.Score < report.Diagnostics.BestScore {
				report.Diagnostics.BestScore = child.Score
			}
		}
	}

	report.BestCandidates = dedupeCandidateScores(report.BestCandidates)
	sort.Slice(report.BestCandidates, func(i, j int) bool {
		if report.BestCandidates[i].Score == report.BestCandidates[j].Score {
			return report.BestCandidates[i].Candidate.Key < report.BestCandidates[j].Candidate.Key
		}
		return report.BestCandidates[i].Score < report.BestCandidates[j].Score
	})
	if len(report.BestCandidates) > options.TopN {
		report.BestCandidates = report.BestCandidates[:options.TopN]
	}

	sort.Slice(report.PartialResults, func(i, j int) bool {
		if report.PartialResults[i].Score == report.PartialResults[j].Score {
			return report.PartialResults[i].Candidate.Key < report.PartialResults[j].Candidate.Key
		}
		return report.PartialResults[i].Score < report.PartialResults[j].Score
	})

	if math.IsInf(report.Diagnostics.BestScore, 1) {
		report.Diagnostics.BestScore = 0
	}

	return report, nil
}

func expandThread(
	thread GrowthThread,
	atoms []ast.Expr,
	samples []common.Sample[float64],
	backend eval.Backend[complex128],
	options MazeOptions,
	seen map[string]bool,
	diagnostics *MazeDiagnostics,
) ([]GrowthThread, []PartialResult, error) {
	var children []GrowthThread
	var partials []PartialResult

	for _, frontier := range thread.Frontiers {
		for _, atom := range atoms {
			for _, candidateExpr := range []struct {
				direction string
				expr      ast.Expr
			}{
				{direction: "left", expr: ast.Apply{Left: cloneExpr(thread.Current.Original), Right: cloneExpr(atom)}},
				{direction: "right", expr: ast.Apply{Left: cloneExpr(atom), Right: cloneExpr(thread.Current.Original)}},
			} {
				candidate := common.NewCandidate(candidateExpr.expr)
				if !common.WithinBounds(candidate.Original, options.Bounds) {
					continue
				}
				if seen[candidate.Key] {
					diagnostics.DuplicateEliminations++
					continue
				}
				seen[candidate.Key] = true
				diagnostics.BranchesExpanded++

				score, err := common.RealMSE(candidate, backend, samples)
				if err != nil {
					diagnostics.BranchesPruned++
					continue
				}
				if !isFinite(score) {
					diagnostics.BranchesPruned++
					continue
				}
				step := ExpansionStep{
					ParentKey: thread.Current.Key,
					Direction: candidateExpr.direction,
					Atom:      atom.String(),
					Score:     score,
					ResultKey: candidate.Key,
					Frontier:  frontier.Name,
					Improved:  score < thread.Score,
				}
				history := append(append([]ExpansionStep(nil), thread.History...), step)
				if score <= options.AcceptThreshold {
					children = append(children, GrowthThread{
						AnchorName: thread.AnchorName,
						Current:    candidate,
						Score:      score,
						Frontiers:  []Frontier{{Name: "root"}},
						History:    history,
						Active:     true,
					})
					continue
				}
				if score <= options.RetainThreshold {
					partials = append(partials, PartialResult{
						AnchorName: thread.AnchorName,
						Candidate:  candidate,
						Score:      score,
						Reason:     "retained_after_validation",
						History:    history,
					})
					continue
				}
				diagnostics.BranchesPruned++
			}
		}
	}

	return children, partials, nil
}

func dedupeCandidateScores(in []CandidateScore) []CandidateScore {
	seen := make(map[string]CandidateScore)
	for _, candidate := range in {
		prev, ok := seen[candidate.Candidate.Key]
		if !ok || candidate.Score < prev.Score {
			seen[candidate.Candidate.Key] = candidate
		}
	}
	out := make([]CandidateScore, 0, len(seen))
	for _, candidate := range seen {
		out = append(out, candidate)
	}
	return out
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func cloneExpr(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case ast.One:
		return ast.One{}
	case ast.Variable:
		return ast.Variable{Name: n.Name}
	case ast.Apply:
		return ast.Apply{
			Left:  cloneExpr(n.Left),
			Right: cloneExpr(n.Right),
		}
	default:
		return nil
	}
}
