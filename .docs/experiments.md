# Experiment Methodology

## Purpose

This document defines how oracle-controlled symbolic-regression experiments
should be designed, executed, and interpreted in this repository.

The goal is controlled recovery, not open-ended discovery.

That means every experiment should begin from a known target law and ask a
specific question:

- can the current EML tooling recover the intended structure from data,
- under explicit search bounds and sampling conditions,
- with results classified in a reproducible way.

This methodology exists so the project can produce empirical results that are
credible enough to support a future companion paper to the underlying EML paper.

## Current Scope

The current experiment system should be aligned with what the codebase actually
supports today:

- raw EML parsing and normalization,
- concept expansion to raw EML,
- real-valued dataset generation from known targets,
- bounded enumerative real search,
- candidate diagnostics,
- deterministic normalized-form comparison,
- proof-friendly export of normalized raw EML artifacts.

The current experiment methodology does not assume:

- open-ended scientific discovery,
- fitted constants or parameter optimization,
- complex-valued oracle experiments as the baseline,
- full semantic equivalence proving,
- or Lean translation as part of the experiment loop.

Those may become future directions, but they are not the current basis for
interpreting results.

## Study Types

Experiments should be grouped into three study types.

### Positive Controls

These are targets the current system should be able to recover under reasonable
search bounds.

Examples:

- `exp`
- `log`
- `sin`
- `sigmoid`

Positive controls establish that the pipeline works at all.

### Negative Controls

These are targets that the current system is not expected to recover under the
chosen search bounds or search mode.

Negative controls are required. They show whether the system fails honestly and
whether diagnostics explain the failure.

### Stretch Controls

These are targets near the edge of current capability.

They should be difficult enough that recovery is uncertain, but not so far
beyond the current search strategy that failure is guaranteed.

Stretch controls help map the boundary between what the current implementation
can recover and what it cannot.

## Recovery Outcome Classes

Every experiment result must be assigned exactly one recovery class.

The classes are:

- `exact_normalized_recovery`
- `concept_equivalent_recovery`
- `approximate_only_recovery`
- `no_recovery`

These labels are mutually exclusive.

### Exact Normalized Recovery

Assign this class when:

- the top-ranked recovered candidate has the exact expected normalized canonical
  key defined by the experiment.

This is the strongest success criterion in the current system.

### Concept-Equivalent Recovery

Assign this class when:

- the top-ranked recovered candidate does not match the primary expected
  canonical key,
- but it does match one of the explicitly allowed equivalent keys declared by
  the experiment.

This class exists because more than one normalized raw EML witness may be
acceptable for a named mathematical target.

### Approximate-Only Recovery

Assign this class when:

- no structural recovery criterion is satisfied,
- but the top-ranked candidate achieves a numeric score at or better than an
  explicit threshold declared by the experiment.

This class must remain separate from structural recovery. A numerically close
candidate is not evidence of exact or concept-level recovery.

### No Recovery

Assign this class when:

- none of the prior criteria are satisfied.

This includes both complete failure and plausible but insufficient near misses.

## Classification Order

Recovery classification must be applied in this order:

1. exact normalized recovery
2. concept-equivalent recovery
3. approximate-only recovery
4. no recovery

This ordering prevents numeric closeness from masking a structural result.

## Partial Recovery Outcome Classes

Maze-mode experiments (`search.mode = "maze_real"`) study fractional recovery:
which labeled pieces of a target law survive search, even when the whole law
is out of reach. The strategic rationale and the layered model these classes
serve are defined in `.docs/partial-recovery-strategy.md`. They use a
separate, mutually exclusive class set:

- `full_law_recovery`
- `snippet_recovery`
- `partial_coverage_recovery`
- `no_recovery`

The whole-formula classes and the partial classes never mix: an experiment
declares one search mode and is classified only against that mode's class set.

### Anchors Are Declared Inputs

Every maze experiment seeds its search from snippets in a committed snippet
artifact. Anchors are therefore part of the experiment's declared conditions,
with full snippet provenance recorded in the result artifact. A partial
recovery claim is always conditional: "given these declared anchors, search
recovered this structure." Maze experiments do not claim unanchored discovery.

### Full Law Recovery

Assign this class when:

- the top-ranked maze candidate has the exact expected normalized canonical
  key declared by the experiment.

This is the maze analogue of exact normalized recovery, and the only maze
class that supports whole-law claims.

### Snippet Recovery

Assign this class when:

- full law recovery is not satisfied,
- but a declared expected snippet canonical key appears among the returned
  top-N candidates.

Snippet recovery deliberately checks the whole returned top-N rather than only
the top rank. Fractional recovery asks which labeled subtrees survive search,
not only which candidate ranks first; N is declared in the spec and all N
candidates are committed in the result artifact, so the criterion stays
auditable.

### Partial Coverage Recovery

Assign this class when:

- no structural criterion is satisfied,
- coverage-aware scoring is enabled in the spec,
- and the top-ranked candidate's best window meets both declared thresholds:
  coverage ratio at or above `min_coverage_ratio` and local error at or below
  `max_local_error`.

This is the partial analogue of approximate-only recovery: a candidate that
explains a declared fraction of the data within a declared error bound, with
no claim about the remainder of the trace. The window, coverage ratio, and
local error are recorded per candidate in the result artifact.

### No Recovery (Maze)

Assign this class when none of the prior criteria are satisfied. As with the
whole-formula classes, this includes honest near misses.

### Partial Classification Order

1. full law recovery (top candidate only)
2. snippet recovery (declared keys against returned top-N)
3. partial coverage recovery (top candidate's window against declared thresholds)
4. no recovery

Structural recovery always outranks coverage-based recovery, mirroring the
whole-formula rule that numeric closeness never masks a structural result.

### Claims Partial Recovery Can Support

- "Under declared anchors and bounds, search recovered the full target law."
- "Under bounds too tight for the full law, search recovered a declared
  labeled snippet of the target."
- "A candidate explains a declared fraction of the trace within a declared
  error bound."

### Claims Partial Recovery Cannot Support

- Any claim of unanchored or open-ended discovery.
- Any claim that a covering window generalizes beyond its declared region.
- Any claim that snippet recovery implies the full law is reachable.

## Required Metadata Per Experiment Run

Every experiment run must record enough metadata to be reproducible and
auditable.

Required fields:

- experiment id
- experiment description
- study type
- target type:
  - named concept
  - raw EML expression
- target identifier or raw expression
- target canonical form where available
- dataset generation method
- variable names
- sample domain
- sample count
- search mode
- search bounds
- scoring function
- expected recovery criterion
- recovery classification result
- code version identifier where available
- timestamp

The run record should also retain:

- search diagnostics,
- top candidates,
- normalized candidate strings,
- canonical keys,
- and dataset provenance.

## Publishable Success And Failure Cases

For the purposes of future empirical writing, the current methodology should
support only restrained claims.

### Publishable Success Cases

The following are credible success cases:

- the system exactly recovers normalized targets on positive controls,
- the system recovers declared concept-equivalent targets when exact normalized
  recovery is not the only acceptable witness,
- diagnostics remain coherent across successful runs,
- and results are reproducible from committed experiment specs.

### Publishable Failure Cases

The following are equally important and should be documented:

- expected failures on negative controls,
- failure on stretch controls beyond the current search boundary,
- failures explained by current search bounds,
- failures explained by documented backend or semantic limitations.

Negative results are valuable if they are reproducible and honestly
characterized.

### Claims The Current Methodology Does Not Support

The current methodology does not justify claims such as:

- general symbolic-regression superiority,
- recovery of arbitrary formulas,
- scientific-law discovery in the open-ended sense,
- or semantic completeness of the current EML implementation.

Any future paper using these experiments must keep its claims tied to
oracle-controlled recovery under explicit search limits.

## Interpretation Rules

When discussing results, use these rules:

- structural recovery is stronger than numeric closeness,
- positive controls and negative controls must both be reported,
- diagnostics must be used to explain failure modes,
- and the search mode and bounds must always be named.

Do not summarize an experiment as “successful” without naming its recovery
class.

Do not treat approximate-only recovery as equivalent to formula recovery.

## Threats To Validity

The current methodology must be interpreted in light of current repository
limitations.

Known threats include:

- bounded enumerative search only,
- no fitted constants or optimizer,
- incomplete branch-sensitive validation for some inverse functions,
- high-precision backend limitations already tracked in `BACKLOG.md`,
- and lack of a full symbolic equivalence prover.

These limitations do not invalidate the experiments, but they constrain what
the results can mean.

## Practical Rule

If a new experiment cannot be classified unambiguously using the categories in
this document, the methodology should be updated before the experiment is used
as evidence.

If a workaround or known limitation affects interpretation of an experiment
result, that limitation should be recorded in `BACKLOG.md` and referenced by
the experiment notes.
