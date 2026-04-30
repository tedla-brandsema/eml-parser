package maze

import (
	"math"
	"sort"

	"eml-parser/ast"
	"eml-parser/eval"
	"eml-parser/search/common"
)

type ThreadStatus string

const (
	ThreadStatusActive          ThreadStatus = "active"
	ThreadStatusRetainedPartial ThreadStatus = "retained_partial"
	ThreadStatusPruned          ThreadStatus = "pruned"
	ThreadStatusCompleted       ThreadStatus = "completed"
)

type Anchor struct {
	Name       string
	Expr       ast.Expr
	Provenance *AnchorProvenance
}

type Frontier struct {
	Index int
	Path  string
	Expr  ast.Expr
	Open  bool
}

type ExpansionStep struct {
	ParentKey    string
	Direction    string
	Atom         string
	Score        float64
	ResultKey    string
	Frontier     string
	FrontierPath string
	Improved     bool
	Improvement  float64
}

type GrowthThread struct {
	AnchorName  string
	Provenance  *AnchorProvenance
	Current     common.Candidate
	Score       float64
	ParentScore float64
	Frontiers   []Frontier
	History     []ExpansionStep
	Status      ThreadStatus
	StopReason  string
}

type PartialResult struct {
	AnchorName    string
	Provenance    *AnchorProvenance
	Candidate     common.Candidate
	Score         float64
	Reason        string
	FrontierCount int
	History       []ExpansionStep
}

type MazeOptions struct {
	Bounds          common.Bounds
	TopN            int
	AcceptThreshold float64
	RetainThreshold float64
	MinImprovement  float64
	Atoms           []ast.Expr
}

type MazeDiagnostics struct {
	AnchorCount                int
	ThreadsSpawned             int
	BranchesExpanded           int
	BranchesPruned             int
	BranchesRetained           int
	BranchesCompleted          int
	DuplicateEliminations      int
	FrontierExpansionsTried    int
	FrontierExpansionsAccepted int
	FrontierExpansionsRejected int
	MaxDepthReached            int
	MaxFrontierCountSeen       int
	BestScore                  float64
}

type CandidateScore struct {
	Candidate  common.Candidate
	Score      float64
	AnchorName string
	Provenance *AnchorProvenance
	History    []ExpansionStep
}

type MazeReport struct {
	BestCandidates []CandidateScore
	PartialResults []PartialResult
	Diagnostics    MazeDiagnostics
}

func MazeRealSearch(fixture common.BenchmarkCase[float64], backend eval.Backend[complex128], anchors []Anchor, options MazeOptions) (MazeReport, error) {
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
	target := common.NewSearchTarget([]string{fixture.TargetKey}, fixture.Samples)
	policy := common.ThresholdRetentionPolicy{
		AcceptThreshold: options.AcceptThreshold,
		RetainThreshold: options.RetainThreshold,
		MinImprovement:  options.MinImprovement,
	}
	return MazeRealSearchWithPolicies(target, backend, anchors, options, common.RealMSEScorer{}, policy)
}

func MazeRealSearchWithPolicies(
	target common.SearchTarget[float64],
	backend eval.Backend[complex128],
	anchors []Anchor,
	options MazeOptions,
	scorer common.Scorer[float64],
	retention common.RetentionPolicy,
) (MazeReport, error) {
	if len(anchors) == 0 {
		return MazeReport{}, nil
	}
	atoms := options.Atoms
	if len(atoms) == 0 {
		atoms = common.AtomicSeeds(target.VariableNames()...)
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
		scored, err := scorer.ScoreCandidate(candidate, backend, target)
		if err != nil || !scored.Finite {
			continue
		}
		threadFrontiers := openFrontiers(candidate.Original)
		thread := GrowthThread{
			AnchorName:  anchor.Name,
			Provenance:  cloneAnchorProvenance(anchor.Provenance),
			Current:     candidate,
			Score:       scored.Primary,
			ParentScore: scored.Primary,
			Frontiers:   threadFrontiers,
			Status:      ThreadStatusActive,
		}
		stack = append(stack, thread)
		seen[candidate.Key] = true
		report.Diagnostics.ThreadsSpawned++
		report.Diagnostics.MaxDepthReached = max(report.Diagnostics.MaxDepthReached, candidate.Stats.TreeDepth)
		report.Diagnostics.MaxFrontierCountSeen = max(report.Diagnostics.MaxFrontierCountSeen, len(threadFrontiers))
		if scored.Primary < report.Diagnostics.BestScore {
			report.Diagnostics.BestScore = scored.Primary
		}
		report.BestCandidates = append(report.BestCandidates, CandidateScore{
			Candidate:  candidate,
			Score:      scored.Primary,
			AnchorName: anchor.Name,
			Provenance: cloneAnchorProvenance(anchor.Provenance),
		})
	}

	for len(stack) > 0 {
		thread := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		children, partials, completed, err := expandThread(thread, atoms, target, backend, options, scorer, retention, seen, &report.Diagnostics)
		if err != nil {
			return MazeReport{}, err
		}
		report.PartialResults = append(report.PartialResults, partials...)
		report.Diagnostics.BranchesRetained += len(partials)
		if completed {
			report.Diagnostics.BranchesCompleted++
		}

		if len(children) == 0 {
			currentScore := common.ScoreResult{Primary: thread.Score, Finite: true}
			parentScore := &common.ScoreResult{Primary: thread.ParentScore, Finite: true}
			outcome := retention.Decide(common.RetentionContext{Parent: parentScore, Current: currentScore})
			if outcome.Decision != common.RetentionPrune && thread.Status != ThreadStatusRetainedPartial {
				report.PartialResults = append(report.PartialResults, PartialResult{
					AnchorName:    thread.AnchorName,
					Provenance:    cloneAnchorProvenance(thread.Provenance),
					Candidate:     thread.Current,
					Score:         thread.Score,
					Reason:        "dead_end",
					FrontierCount: len(thread.Frontiers),
					History:       append([]ExpansionStep(nil), thread.History...),
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
				Provenance: cloneAnchorProvenance(child.Provenance),
				History:    append([]ExpansionStep(nil), child.History...),
			})
			report.Diagnostics.MaxDepthReached = max(report.Diagnostics.MaxDepthReached, child.Current.Stats.TreeDepth)
			report.Diagnostics.MaxFrontierCountSeen = max(report.Diagnostics.MaxFrontierCountSeen, len(child.Frontiers))
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
	target common.SearchTarget[float64],
	backend eval.Backend[complex128],
	options MazeOptions,
	scorer common.Scorer[float64],
	retention common.RetentionPolicy,
	seen map[string]bool,
	diagnostics *MazeDiagnostics,
) ([]GrowthThread, []PartialResult, bool, error) {
	var children []GrowthThread
	var partials []PartialResult
	var completed = true

	for _, frontier := range thread.Frontiers {
		if !frontier.Open {
			continue
		}
		for _, atom := range atoms {
			for _, shape := range []struct {
				direction string
				expr      ast.Expr
			}{
				{direction: "left", expr: ast.Apply{Left: cloneExpr(frontier.Expr), Right: cloneExpr(atom)}},
				{direction: "right", expr: ast.Apply{Left: cloneExpr(atom), Right: cloneExpr(frontier.Expr)}},
			} {
				diagnostics.FrontierExpansionsTried++
				replaced, err := common.ReplaceSubtree(thread.Current.Original, frontier.Index, shape.expr)
				if err != nil {
					diagnostics.FrontierExpansionsRejected++
					continue
				}
				candidate := common.NewCandidate(replaced)
				if !common.WithinBounds(candidate.Original, options.Bounds) {
					diagnostics.FrontierExpansionsRejected++
					continue
				}
				if seen[candidate.Key] {
					diagnostics.DuplicateEliminations++
					diagnostics.FrontierExpansionsRejected++
					continue
				}
				seen[candidate.Key] = true
				diagnostics.BranchesExpanded++

				scored, err := scorer.ScoreCandidate(candidate, backend, target)
				if err != nil || !scored.Finite {
					diagnostics.BranchesPruned++
					diagnostics.FrontierExpansionsRejected++
					continue
				}
				parentScore := common.ScoreResult{Primary: thread.Score, Finite: true}
				outcome := retention.Decide(common.RetentionContext{
					Parent:  &parentScore,
					Current: scored,
				})
				improvement := thread.Score - scored.Primary
				step := ExpansionStep{
					ParentKey:    thread.Current.Key,
					Direction:    shape.direction,
					Atom:         atom.String(),
					Score:        scored.Primary,
					ResultKey:    candidate.Key,
					Frontier:     frontier.Path,
					FrontierPath: frontier.Path,
					Improved:     improvement > 0,
					Improvement:  improvement,
				}
				history := append(append([]ExpansionStep(nil), thread.History...), step)

				switch outcome.Decision {
				case common.RetentionContinue:
					completed = false
					diagnostics.FrontierExpansionsAccepted++
					children = append(children, GrowthThread{
						AnchorName:  thread.AnchorName,
						Provenance:  cloneAnchorProvenance(thread.Provenance),
						Current:     candidate,
						Score:       scored.Primary,
						ParentScore: thread.Score,
						Frontiers:   openFrontiers(candidate.Original),
						History:     history,
						Status:      ThreadStatusActive,
					})
				case common.RetentionRetainPartial:
					diagnostics.FrontierExpansionsRejected++
					partials = append(partials, PartialResult{
						AnchorName:    thread.AnchorName,
						Provenance:    cloneAnchorProvenance(thread.Provenance),
						Candidate:     candidate,
						Score:         scored.Primary,
						Reason:        outcome.Reason,
						FrontierCount: len(openFrontiers(candidate.Original)),
						History:       history,
					})
				case common.RetentionPrune:
					diagnostics.BranchesPruned++
					diagnostics.FrontierExpansionsRejected++
				}
			}
		}
	}

	return children, partials, completed, nil
}

func openFrontiers(expr ast.Expr) []Frontier {
	type subtreeRef struct {
		Index int
		Path  string
		Expr  ast.Expr
	}
	var refs []subtreeRef
	index := 0
	var walk func(ast.Expr, string)
	walk = func(node ast.Expr, path string) {
		refs = append(refs, subtreeRef{
			Index: index,
			Path:  path,
			Expr:  cloneExpr(node),
		})
		index++
		if app, ok := node.(ast.Apply); ok {
			walk(app.Left, path+".L")
			walk(app.Right, path+".R")
		}
	}
	walk(expr, "root")

	out := make([]Frontier, 0, len(refs))
	for _, ref := range refs {
		out = append(out, Frontier{
			Index: ref.Index,
			Path:  ref.Path,
			Expr:  ref.Expr,
			Open:  true,
		})
	}
	return out
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

func cloneAnchorProvenance(in *AnchorProvenance) *AnchorProvenance {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
