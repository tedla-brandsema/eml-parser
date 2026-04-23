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

## Experiment Sequence

### 14. Experiment Methodology Doc

Status: implemented

Create a dedicated experiment-methodology document for oracle-controlled symbolic-regression studies.

Targets:

- add `.docs/experiments.md`
- define the purpose of oracle-controlled experiments as controlled recovery, not open-ended discovery
- define study types:
  - positive controls
  - negative controls
  - stretch controls
- define recovery outcome classes:
  - exact normalized recovery
  - concept-equivalent recovery
  - approximate-only recovery
  - no recovery
- define the metadata every experiment run must record:
  - experiment id
  - target type
  - target identifier or raw expression
  - dataset generation method
  - sample domain and count
  - search bounds
  - scoring function
  - recovery criterion
  - code version identifier where available
- define what counts as a publishable success or failure case
- make the recovery labels mutually exclusive and unambiguous

Reason:

- if this project is going to support an empirical companion paper, experiment interpretation needs to be fixed before large amounts of data are generated
- ad hoc experiment notes will make later analysis unreliable

### 15. Experiment Directory Layout

Status: implemented

Create a stable on-disk layout for experiment inputs and outputs.

Targets:

- add an `experiments/` directory at the project root
- define subdirectories for:
  - `experiments/specs/`
  - `experiments/datasets/`
  - `experiments/results/`
  - `experiments/reports/`
- document:
  - which files are source-of-truth inputs
  - which files are generated artifacts
  - which outputs belong in version control
- use deterministic output naming:
  - `<experiment-id>.json`
  - `<suite-id>.json`
  - `<suite-id>.md`
- keep the layout simple enough for both CLI-driven runs and later paper-oriented aggregation

Reason:

- experiment data needs a predictable home before repeated runs start producing artifacts

### 16. Oracle Experiment Schema

Status: implemented

Define a machine-readable experiment schema for oracle-controlled recovery tasks.

Targets:

- define one JSON experiment spec format for `experiments/specs/*.json`
- required top-level fields:
  - `id`
  - `description`
  - `target`
  - `dataset`
  - `search`
  - `recovery`
- support exactly two target forms:
  - named concept target
  - raw EML string target
- support exactly two v1 dataset modes:
  - explicit real grid
  - explicit list of real sample points
- `search` must include:
  - search mode fixed to current enumerative real search
  - bounds with `max_depth` and `max_nodes`
  - `top_n`
- `recovery` must include:
  - expected recovery class
  - expected canonical key when exact recovery is required
  - optional allowed concept-equivalent keys
- keep the schema downstream of the concept dictionary and raw search layers

Reason:

- experiments should be declarative and reproducible rather than embedded implicitly in test code

### 17. Oracle Dataset Generator

Status: implemented

Build a reusable dataset generator for controlled recovery experiments.

Targets:

- generate datasets from known target concepts or raw expressions
- support real-valued positive controls first
- support negative controls where current bounds should fail honestly
- support deterministic real-domain generation:
  - evenly spaced grid over `[min, max]`
  - explicit real sample list
- use the existing evaluation layer to generate target values
- write generated datasets to `experiments/datasets/<experiment-id>.json`
- include dataset metadata in the artifact:
  - target
  - variable name
  - domain
  - sample count
  - generator mode

Reason:

- symbolic-regression claims will only be credible if the data source and target law are both explicit and reproducible

### 18. Oracle Search Harness

Status: implemented

Build a dedicated experiment runner around the existing search skeleton.

Targets:

- load experiment definitions
- generate or load the corresponding dataset
- run bounded search
- record diagnostics, recovered candidates, and recovery classification
- convert datasets into the sample format expected by the current search package
- write one result artifact per run to `experiments/results/<experiment-id>.json`
- keep the first harness aligned with the current enumerative search loop instead of inventing a more advanced search strategy
- keep the first harness real-valued only

Reason:

- benchmark fixtures exist, but they are not yet a proper experiment system

### 19. Recovery Classification

Status: implemented

Add explicit recovery classification for experiment results.

Targets:

- distinguish:
  - exact normalized recovery
  - concept-equivalent recovery
  - approximate-only recovery
  - no recovery
- classify in this order:
  1. exact normalized recovery
  2. concept-equivalent recovery
  3. approximate-only recovery
  4. no recovery
- define exact normalized recovery as top candidate canonical key equals expected canonical key
- define concept-equivalent recovery as candidate canonical key matches one of the allowed equivalent keys from the experiment spec
- define approximate-only recovery as no structural match but best score passes a numeric threshold declared in the experiment spec
- keep numeric-only recovery separate from structural recovery

Reason:

- for paper-worthy results, “found something close” is too vague

### 20. Experiment Result Recording

Status: implemented

Record experiment runs in a stable machine-readable format.

Targets:

- persist:
  - experiment metadata
  - dataset metadata
  - search options
  - recovered candidates
  - scores
  - diagnostics
  - recovery classification
  - target canonical form
  - timestamps
  - code version identifiers where practical
- prefer JSON artifacts that can later be aggregated
- include both normalized candidate strings and canonical keys
- include enough provenance to trace:
  - which experiment spec produced the run
  - which dataset was used
  - which code version produced the output
- keep the format compatible with the current formalization/export direction

Reason:

- results should be auditable and reusable for later analysis, not just printed to the terminal

### 21. Experiment Reporting

Status: implemented

Add lightweight summary reporting for experiment suites.

Targets:

- summarize success and failure counts
- summarize recovery classes across a suite
- summarize diagnostics such as candidate counts and rejection rates
- surface the top recovered expressions and their normalized forms
- produce:
  - machine-readable suite summary JSON
  - human-readable suite summary Markdown
- summarize by:
  - recovery class
  - target family
  - aggregate diagnostics ranges
- keep the first reporting pass textual and machine-readable before attempting polished paper tables or charts

Reason:

- once multiple oracle experiments exist, we need a way to see patterns rather than isolated runs

### 22. Initial Oracle Suite

Create the first paper-oriented oracle experiment suite.

Targets:

- include exact-recovery controls for current library concepts such as:
  - `exp`
  - `log`
  - `sin`
  - `sigmoid`
- include small composite controls built from known concepts:
  - one additive composite
  - one multiplicative or nested composite
- include at least two negative controls that current bounds should fail to recover
- define for each experiment:
  - target
  - dataset
  - search bounds
  - expected recovery class
- keep the first suite small enough to inspect manually

Reason:

- this will establish the first empirical baseline for what the current codebase can and cannot recover

### 23. Paper Readiness Notes

Document how experiment results map to future empirical writing.

Targets:

- record what claims the current experiments can support
- record what claims they cannot yet support
- document current threats to validity:
  - limited search strategy
  - no fitted constants or optimizer
  - high-precision backend limitations
  - incomplete branch-sensitive validation
- keep this documentation in `.docs` rather than burying it in code comments
- tie claims to observed recovery classes rather than informal impressions

Reason:

- if the project is moving toward a paper, limits and caveats should be documented as rigorously as positive findings

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
