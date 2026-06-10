# Maze Oracle Suite

This suite establishes the first empirical baseline for fractional recovery:
snippet-anchored maze search classified with the partial recovery classes
defined in `.docs/experiments.md`.

All experiments seed their anchors from the committed curated snippet corpus
(`artifacts/snippets/`, regenerated deterministically by
`emltool gen-snippet-datasets` when missing).

Included experiments:

- `maze_oracle_exp3_full_from_exp2.json`
  - full-law positive control: the `exp2` snippet anchor grows into the whole
    `exp(exp(exp(x)))` law
- `maze_oracle_exp3_snippet_from_exp1.json`
  - snippet-recovery positive control: under bounds too tight for the full
    law, the `exp1` anchor still recovers the labeled `exp2` snippet in top-N
- `maze_oracle_sinh_partial_coverage.json`
  - partial-coverage stretch control: with coverage-aware scoring, `x` is
    recovered as a partial law for `sinh(x)` over the small-x region
- `maze_oracle_sigmoid_negative.json`
  - negative control: exponential anchors fail honestly against `sigmoid`
    under strict declared coverage criteria

Rationale:

- the full-law control shows anchored growth can complete a law that bounded
  enumerative search alone would have to rediscover from atoms
- the snippet control shows labeled subtrees are recoverable as first-class
  results when the whole law is out of reach
- the coverage control shows fractional fits are classified by declared
  thresholds, not impressions
- the negative control shows coverage-aware classification does not inflate
  weak fits into recoveries

Current intended reporting flow:

1. run each spec with `emltool run-experiment`
2. aggregate the result files with `emltool report-suite`

Example:

```bash
go run ./cmd/emltool run-experiment experiments/specs/maze_oracle_exp3_full_from_exp2.json
go run ./cmd/emltool run-experiment experiments/specs/maze_oracle_exp3_snippet_from_exp1.json
go run ./cmd/emltool run-experiment experiments/specs/maze_oracle_sinh_partial_coverage.json
go run ./cmd/emltool run-experiment experiments/specs/maze_oracle_sigmoid_negative.json
go run ./cmd/emltool report-suite maze_oracle_suite \
  experiments/results/maze_oracle_exp3_full_from_exp2.json \
  experiments/results/maze_oracle_exp3_snippet_from_exp1.json \
  experiments/results/maze_oracle_sinh_partial_coverage.json \
  experiments/results/maze_oracle_sigmoid_negative.json
```
