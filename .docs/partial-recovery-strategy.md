# Partial Recovery Strategy

## Purpose

This document codifies the strategic bet that partial recovery — anchored,
layered, locally valid explanations of data — is the most valuable long-term
approach in this project, and defines the layered model that future partial
recovery, junction analysis, and assembly work must follow.

It extends, and does not replace, the experiment methodology in
`.docs/experiments.md`. The partial recovery classes defined there remain the
classification vocabulary; this document defines what partial recovery is
*for* and where it is going.

## Why Partial Recovery Is The Primary Bet

Whole-formula recovery can only confirm rediscovery: it requires that the
target law already be expressible, reachable, and known well enough to declare
a canonical key. That makes it the right instrument for oracle-controlled
calibration, and the wrong instrument for the actual research frontier.

Bleeding-edge mathematical and scientific datasets, by definition, do not fit
known closed-form laws. There is no whole law to recover. What such data does
contain is local structure: asymptotic behavior, limiting regimes, regions
where a simple law holds and regions where it visibly stops holding. This is
exactly what working scientists publish about unknown functions — series
expansions, dominant balances, boundary behavior.

The first committed partial-coverage experiment already demonstrates the
shape of this: recovering `x` as the small-x law of `sinh(x)` is recovering
the first term of its Taylor expansion. That is not a degraded whole-formula
result. It is the correct kind of finding for data whose global law is out of
reach.

There is an established mathematical precedent for this approach: matched
asymptotic expansions in boundary-layer theory, where different local
solutions are derived in different regimes and reconciled in the overlap
regions where both are valid. The layered model below is that idea,
generalized to anchored search.

## The Layered Model

The working mental model: a dataset is a base layer. Every anchor found in
the data spawns a transparent layer above it. Each layer grows outward from
its anchor independently. When all layers have reached their natural end,
the structure of interest is no longer any single layer but the geometry
between layers — where they touch, where they overlap, and where they
disagree.

### Layer

A layer is the unit of partial explanation. One layer is:

- one anchor, with full provenance (snippet origin or future anchor sources)
- one candidate expression grown from that anchor
- the set of sample windows the expression explains within declared tolerance
- the error profile across those windows
- the boundaries where the explanation stops holding

A layer is always conditional on its anchor. Layers never claim unanchored
discovery.

### Natural End Of A Layer

A layer's extent must ultimately be a measured property of the data, not only
a configured threshold. The current implementation stops growth on declared
accept/retain thresholds and bounds; that is acceptable for oracle-controlled
calibration but is not the end state. The principled boundary is where local
error inflects — where the law stops holding. That boundary is itself a
finding: it marks structure in the data, and it must be recorded, not
discarded.

### Junctions Between Layers

When two layers touch or overlap, the relation between them must be
classified, not collapsed. Three junction relations are distinguished:

- **Agreement** — both expressions explain the overlap within declared
  tolerance and are behaviorally interchangeable there. This is observed,
  regional equivalence: equivalence-family material discovered from data
  rather than declared by construction.
- **Complementarity** — different laws on adjacent or lightly overlapping
  windows, with compatible values at the seam. This is assembly material:
  candidate pieces of a larger explanation.
- **Disagreement** — both expressions fit the shared region within tolerance
  but are structurally and behaviorally different elsewhere. This is the most
  informative relation: the same data interpreted differently because the
  layers grew from different seeds. Disagreement regions identify exactly
  where more data or deeper search would discriminate between hypotheses.

### Overlap As A Training-Data Factory

Agreement junctions close a loop with the ML strategy. The equivalence-family
corpus currently contains only relations declared by construction. Layer
overlaps generate equivalence candidates from data — observed regional
equivalences with recorded windows, tolerances, and anchor provenance. These
feed the same artifact contract that the `family_match` task already
consumes, and they are precisely the kind of equivalence-aware exposure
`.docs/ml-strategy.md` calls for.

## Multi-Window Coverage Is Required

The current partial-coverage scorer retains only the single best contiguous
window per candidate. For the layered model this is insufficient, and the
limitation is structural, not cosmetic: usefulness does not live in one best
window.

- A law that genuinely holds in two disjoint regions of the domain cannot be
  represented at all.
- Layer extents are truncated to one interval, so junction analysis would
  systematically miss touchpoints and overlaps.
- Boundary detection requires the full per-sample error profile, not one
  window summary.

Multi-window coverage — a set of disjoint explained windows per candidate,
each with its own local error — is therefore a prerequisite for layers,
junctions, and assembly, and is the highest-priority implementation gap in
this strategy.

## Claim Discipline

The honest-classification rules from `.docs/experiments.md` extend to layers
and assembly:

- A layer claim is always conditional: "given this declared anchor, this
  expression explains these windows within these tolerances."
- A covering window never generalizes beyond its declared region.
- An assembled patchwork of layers is never classified as a law. Assembly
  results get their own outcome classes; they do not promote into whole-law
  recovery classes.
- Junction classifications use declared tolerances, recorded in specs and
  result artifacts, exactly as recovery classes do today.
- Combinatorial analysis over layers must stay bounded and declared: pairwise
  junction analysis first, k-way assembly only under explicit bounds, with
  deduplication by canonical key. Unbounded patchwork search is curve fitting
  with extra steps and is out of scope.

## Relation To Existing Subsystems

- **Maze search** is the layer-growing engine: growth threads already carry
  anchor provenance, history, and retention status.
- **Snippet artifacts** are the current anchor source; future anchor sources
  (including ML-proposed anchors) must arrive with equivalent provenance.
- **Whole-formula oracle experiments** remain the calibration baseline and
  regression guard; they are not deprecated by this strategy.
- **Equivalence families and `ml/`** are the consumers of agreement
  junctions; assembly experiments are the consumers of complementarity
  junctions.

## What This Strategy Does Not Claim

- It does not claim anchored layering finds laws that search alone cannot
  express; bounds and the EML substrate still constrain everything.
- It does not claim overlap disagreement resolves itself; disagreement is
  surfaced for discrimination, not arbitrated automatically.
- It does not replace oracle control. Every new mechanism (multi-window
  scoring, junction analysis, assembly) must arrive with its own
  oracle-controlled positive, negative, and stretch experiments before being
  used on unknown data.
