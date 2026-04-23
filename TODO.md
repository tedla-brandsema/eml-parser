# TODO

## Near-Term Sequence

### 1. Tooling: Expanded Tree Stats

Status: implemented

Add concept-analysis tooling for expanded raw EML trees.

Target:

- add `emltool stats <concept>`

Output should include:

- expanded node count
- tree depth
- leaf terminal count
- direct dependency count
- transitive dependency count

Reason:

- we already have `show` for concept-layer structure
- we already have `expand` for fully reduced raw EML
- the next practical question is how expensive a concept becomes after full expansion

This should be implemented without changing the raw parser or evaluator model.

### 2. Raw EML Normalization

Status: implemented

Add normalization and simplification passes for expanded raw EML trees.

Initial goals:

- structural cleanup of expanded trees
- elimination of obvious redundancies where safe
- stable output for repeated expansion of equivalent concepts

Reason:

- once we can inspect expanded size and depth, we need a way to reduce them
- normalization will matter for caching, comparison, symbolic regression, and later proof/export work

This should operate on raw EML AST only.

### 3. Expansion Caching

Status: implemented

Cache expanded raw EML trees for named concepts.

Initial goals:

- avoid repeated recursive expansion for the same concept
- keep caching separate from concept registration logic
- preserve correctness when concept definitions are immutable

Reason:

- some concepts now expand through long dependency chains
- future tooling and symbolic workflows will repeatedly request the same expanded forms

### 4. Standard Library Growth

Continue extending the concept dictionary carefully.

Next candidates:

- inverse hyperbolic functions
- inverse trigonometric functions
- logistic / sigmoid-style functions
- additional constants and scientific-calculator primitives from the paper

Constraint:

- keep parser grammar unchanged
- add only concept-layer definitions
- prefer grounded, compositional definitions over ad hoc shortcuts

### 5. Search / Regression Preparation

Status: implemented

Prepare the codebase for symbolic-regression workflows over raw EML trees.

Initial goals:

- define search-oriented tree utilities
- make normalized expanded trees easy to measure and compare
- keep concept-layer compilation separate from raw search space operations

Reason:

- the long-term objective is not only evaluation, but generation and discovery of EML trees from data

## Next Sequence

### 6. Standard Library Growth

Status: implemented

Continue extending the concept dictionary carefully.

Targets:

- inverse hyperbolic functions
- inverse trigonometric functions
- logistic / sigmoid-style functions
- additional constants and scientific-calculator primitives from the paper

Constraints:

- keep parser grammar unchanged
- add only concept-layer definitions
- prefer grounded, compositional definitions over ad hoc shortcuts
- keep new concepts recursively expandable to raw EML

### 7. Standard Library Validation

Status: implemented

Strengthen validation around the growing concept library.

Targets:

- identity tests for new concepts against `complex128`
- high-precision validation for representative concepts
- branch-sensitive tests for inverse functions
- normalization regression tests for newly expanded forms
- expansion-size checks for especially expensive concepts

Reason:

- concept growth without stronger validation will make the library hard to trust
- inverse and branch-sensitive functions are where semantic mistakes are most likely

### 8. Concept Library Tooling

Status: implemented

Expand the existing dictionary tooling for better inspection.

Targets:

- combined view of concept-layer definition and expanded raw EML
- dependency-chain / call-stack reporting for a concept
- normalized-vs-expanded comparison output
- size and depth deltas before and after normalization

Reason:

- the library is now large enough that structural inspection matters
- we need visibility into how much complexity each named concept introduces

### 9. Search Space Utilities

Status: implemented

Add raw-tree helpers aimed at candidate generation.

Targets:

- subtree replacement helpers
- tree mutation / rewrite helpers
- depth-bounded and node-bounded construction helpers
- deduplication by normalized canonical key

Constraints:

- operate on raw EML AST only
- keep concept expansion separate from search-space manipulation

### 10. Dataset / Benchmark Layer

Status: implemented

Add small, reusable regression datasets and scoring fixtures.

Targets:

- sample-set helpers for regression experiments
- benchmark fixtures for known target functions
- reproducible evaluation cases for standard-library concepts
- real-only and complex-valued benchmark coverage

Reason:

- search work needs stable targets and comparable scoring cases
- benchmark fixtures should exist before the first real search loop

### 11. First Search Skeleton

Status: implemented

Build the first minimal search workflow on top of the preparation work.

Targets:

- minimal enumerative or mutation-based candidate generator
- candidate scoring over sample sets
- ranking and deduplication by normalized key
- simple inspectable output for top candidates

Constraints:

- keep the first search loop small and transparent
- do not over-engineer optimization before baseline search behavior is visible

### 12. Search Diagnostics

Status: implemented

Add observability around candidate generation and scoring.

Targets:

- candidate counts
- rejection reasons
- normalization hit rates
- score distributions
- top-candidate summaries

Reason:

- search will be difficult to debug without visibility into what is being generated and discarded

### 13. Formalization Bridge

Status: implemented

Prepare normalized raw EML trees for downstream formal workflows.

Targets:

- export of normalized raw EML into a proof-friendly intermediate form
- concept-to-raw provenance retained where possible
- groundwork for later Lean-oriented translation

Constraint:

- keep this downstream of normalization and search, not mixed into the parser or concept registry layers

## Architectural Rules

- The raw parser stays minimal: `1`, variables, `eml(left, right)`.
- Named mathematical concepts belong in the concept dictionary, not in the parser grammar.
- Tooling should inspect and expand concepts without broadening the raw language.
- Normalization should operate on raw EML AST after expansion.
- New work should preserve the split between:
  - raw EML substrate,
  - concept dictionary,
  - expansion,
  - evaluation,
  - tooling.
