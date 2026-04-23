# Paper Readiness Notes

## Purpose

This document records what the current experiment stack can support as
empirical evidence, what it cannot yet support, and which current limitations
must be disclosed if the project is written up as a paper.

The goal is to keep claims aligned with the actual experiment system rather than
with the broader ambitions of the repository.

## What The Current Experiments Can Support

At the current stage, the project can support claims of the following kind:

- the repository can run reproducible oracle-controlled recovery experiments
  over raw EML search candidates,
- the current bounded enumerative real search can exactly recover some small
  target laws under explicit search bounds,
- the current tooling can distinguish exact structural recovery from
  approximate-only matches and from outright failure,
- result artifacts, diagnostics, and suite summaries are reproducible from
  committed experiment specs.

These claims are strongest when tied directly to observed recovery classes:

- `exact_normalized_recovery`
  - strongest current success signal
- `concept_equivalent_recovery`
  - acceptable structured recovery when the spec explicitly allows it
- `approximate_only_recovery`
  - evidence of numeric closeness only, not structural recovery
- `no_recovery`
  - honest failure under the current search budget and implementation

The current initial oracle suite supports restrained statements such as:

- `exp` and `log` are recoverable under the current bounded search when the
  target and search budget are chosen appropriately,
- at least one small nested composite target is also recoverable,
- larger-library functions such as `sin` and `sigmoid` are not yet exact
  controls under the present search regime,
- additive composites beyond the small exact regime already expose a meaningful
  failure boundary.

## What The Current Experiments Cannot Yet Support

The current experiment stack does not justify claims such as:

- general symbolic-regression superiority,
- recovery of arbitrary elementary functions,
- open-ended scientific-law discovery,
- semantic completeness of the current EML implementation,
- or practical competitiveness with mature symbolic-regression systems.

It also does not justify claims that:

- the current search engine is near-optimal,
- the current concept dictionary is semantically complete,
- all mathematically intended identities are branch-correct in the complex
  plane,
- or high-precision evaluation is already robust enough to serve as a final
  trusted oracle for the full library.

Any future empirical writing should avoid sliding from:

- "the system can recover these oracle targets under these bounds"

to:

- "the system can recover elementary functions in general."

The current evidence does not support that broader jump.

## Current Threats To Validity

The following threats should be disclosed explicitly in any write-up based on
the present codebase.

### 1. Limited Search Strategy

The current search engine is:

- bounded,
- enumerative,
- real-valued,
- and structurally small-scale.

This means failure may reflect search limitations rather than a fundamental
limitation of the EML representation itself.

### 2. No Fitted Constants Or Optimizer

The current experiment path does not perform:

- constant fitting,
- parameter optimization,
- gradient-based refinement,
- or hybrid symbolic-numeric search.

That significantly narrows the class of targets the system can recover and must
be acknowledged whenever negative results are discussed.

### 3. Branch-Sensitive Validation Is Incomplete

As documented in `BACKLOG.md`, some inverse-function branch-sensitive cases do
not yet align fully with the intended principal-branch semantics.

This limits how strongly the current experiments can claim semantic correctness
for the broader concept library, especially beyond the real-valued recovery path
used in the oracle suite.

### 4. High-Precision Backend Limitations

As documented in `BACKLOG.md`, the current high-precision backend still has
exact-zero and compositional edge cases.

That means the present experiments should not be framed as relying on a fully
trusted arbitrary-precision oracle across the whole concept library.

### 5. Oracle Suite Coverage Is Still Narrow

The current oracle suite is deliberately small.

That is useful for inspectability, but it means:

- the evidence is still baseline evidence,
- the suite is not yet broad enough to characterize the full reachable space,
- and the current paper-readiness posture is exploratory rather than
  comprehensive.

## How To Phrase Claims Responsibly

Prefer formulations like:

- "Under bounded enumerative real search, the system exactly recovered these
  targets..."
- "The current oracle suite shows a recovery boundary between these small exact
  controls and these current-boundary failures..."
- "Negative controls and larger targets currently fail honestly under the
  present search budget..."

Avoid formulations like:

- "The system discovers elementary functions,"
- "The tool solves symbolic regression,"
- or "The representation is empirically validated in general."

Those statements outrun the present evidence.

## Current Paper Position

If the project were written up today, the strongest credible framing would be:

- an empirical companion effort around a code-oriented EML toolchain,
- showing that a minimal raw EML substrate plus concept expansion plus bounded
  search can support reproducible oracle recovery experiments,
- while also identifying the current recovery boundary and the major missing
  pieces needed for broader claims.

That is a credible contribution.

It is narrower than a full symbolic-regression paper, but it is much more
defensible.

## Practical Rule

When new experiment results are added, update this document if they materially
change:

- what the project can credibly claim,
- what remains outside current evidence,
- or which threats to validity dominate interpretation.
