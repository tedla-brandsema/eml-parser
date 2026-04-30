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

Status: implemented

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

Status: implemented

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

## Search Improvement Sequence

Status: deferred pending the monorepo and equivalence-learning pivot

The items below still matter, but they are no longer the highest-priority next steps.
Current priority is synthetic data generation, equivalence families, snippet-level
artifacts, and the Python `ml/` subproject that will consume them.

### 24. Search Result Hygiene

Add stricter handling for scored candidates so exact-recovery work is not polluted by unusable results.

Targets:

- filter or explicitly quarantine non-finite scores before ranking
- distinguish finite scored candidates from unusable scored candidates in diagnostics
- make best / worst / mean score reporting robust when `+Inf` or `NaN` appears
- preserve JSON-safe score encoding in experiment artifacts
- add regression tests for finite-vs-nonfinite candidate handling

Reason:

- current search output still surfaces non-finite scores
- exact-recovery work needs cleaner ranking and cleaner diagnostics before more advanced search logic is added

### 25. Layered Enumeration

Replace single-pass bounded closure enumeration with a depth-aware layered search.

Targets:

- generate candidates by exact depth or exact node layer rather than one full closure bucket
- record per-layer candidate counts and per-layer best scores
- stop early when exact recovery is achieved
- keep the first implementation deterministic and bounded
- expose layer diagnostics in the existing search report

Reason:

- current search expands everything inside one bound bucket
- layered search will show where recovery first becomes possible and avoid unnecessary work once the target is already found

### 26. Semantic Pruning

Add pruning based on observed behavior over the oracle dataset, not only on structure.

Targets:

- evaluate candidates on the sample set during search
- deduplicate or discard candidates whose sampled outputs are identical or effectively identical
- add configurable tolerance for real-valued semantic dedupe
- retain structural canonical keys separately from semantic signatures
- record semantic-pruning counts in diagnostics

Reason:

- structural dedupe alone leaves many behaviorally redundant candidates
- exact recovery becomes harder when the frontier is crowded with semantically equivalent junk

### 27. Trivial Expression Pruning

Remove low-information candidate families early without pruning away known exact winners.

Targets:

- detect and discard candidates that collapse to constants or near-constants on the full sample set when that makes them obviously unhelpful
- discard exact duplicate variable or constant families already represented by smaller candidates
- keep pruning rules conservative and sample-driven
- document every pruning rule and test it against current exact controls

Reason:

- current failures still rank trivial expressions highly for harder targets
- those baselines are useful, but they should not dominate the frontier indefinitely

### 28. Beam Search Skeleton

Add a second search mode that can go deeper without full exhaustive expansion.

Targets:

- introduce a deterministic beam search over raw EML candidates
- seed from the same atomic basis as current enumerative search
- expand only the top `k` frontier candidates per layer by score
- keep beam width, max depth, and max nodes explicit in options
- preserve the existing enumerative search path as a baseline mode
- compare beam search against enumerative search on the committed oracle suite

Reason:

- exact recovery of current failures likely requires deeper search, but full enumeration will scale poorly
- beam search is the smallest meaningful next step beyond pure exhaustive bounded enumeration

### 29. Mutation / Refinement Search

Use local refinement around promising candidates instead of always rebuilding from atoms.

Targets:

- add a mutation-driven search pass using the existing subtree replacement helpers
- mutate top-ranked candidates from earlier layers or earlier search modes
- support replacement by small seed expressions and previously promising subtrees
- deduplicate mutated candidates by structural and semantic signatures
- record mutation-origin diagnostics so recoveries can be traced back to their parent candidates

Reason:

- the repo already has mutation primitives, but they are not part of the actual search loop
- harder recoveries may require local refinement around near-miss candidates

### 30. Search Mode Integration In Experiments

Make the experiment schema and harness able to compare improved search modes reproducibly.

Targets:

- extend experiment search config to support:
  - `enumerative_real`
  - `beam_real`
  - `enumerative_plus_mutation`
- keep mode validation strict
- preserve backward compatibility for existing committed oracle specs
- record search-mode-specific diagnostics in result artifacts
- add experiment tests proving that old specs still run unchanged

Reason:

- search improvements only matter if the oracle framework can compare them reproducibly

### 31. Oracle Boundary Re-baselining

Use the improved search modes to move the exact-recovery boundary deliberately.

Targets:

- rerun the committed oracle suite after each major search improvement
- record which current negative or stretch controls move to:
  - exact normalized recovery
  - approximate-only recovery
  - still no recovery
- treat `sin` and `sigmoid` as the first named boundary targets
- add at least one search-regression test proving an improved mode strictly dominates the old mode on a current negative or stretch control

Reason:

- the point of search improvement is not abstract elegance
- it is to move concrete oracle targets from failure into exact recovery

### 32. Search Improvement Readout

Document the evolving search strategy with the same rigor as the experiment methodology.

Targets:

- add `.docs/search-strategy.md`
- explain:
  - current search modes
  - why each mode exists
  - what oracle boundary it moved
  - what it still does not solve
- tie every claimed improvement to oracle-suite evidence rather than anecdote

Reason:

- once multiple search modes exist, the repo needs a stable explanation of their role and current limits

## Monorepo And Equivalence Learning Sequence

### 33. Monorepo Layout And Contracts

Status: implemented

Codify the repository as a monorepo with a stable boundary between symbolic generation and ML experimentation.

Targets:

- define the intended top-level split between the Go core and a future Python `ml/` subproject
- document the artifact contract between them:
  - dataset specs
  - generated datasets
  - equivalence-family corpora
  - experiment results
- keep the Go side as the source of truth for symbolic semantics
- keep Python as a consumer of generated artifacts rather than a second symbolic engine

Reason:

- the project direction has shifted from search-only hardening toward synthetic-data and equivalence-aware ML
- the split needs to be explicit before code starts accreting in the wrong places

### 34. Python ML Subproject Scaffold

Status: implemented

Create the initial Python `ml/` area for experiments over Go-generated artifacts.

Targets:

- add an `ml/` subproject with a minimal environment and README
- define how it reads datasets and corpora generated by the Go side
- define a first experiment entrypoint for offline training or evaluation runs
- keep the first scaffold small and artifact-driven

Reason:

- the repo needs a clear place for ML work before training logic or notebooks appear ad hoc

### 35. Synthetic Tree Family Generator

Use the Go core to generate controlled families of raw EML trees and concept expansions.

Targets:

- generate datasets from single raw trees
- generate datasets from named concepts expanded to raw EML
- support seeded deterministic generation
- record the originating raw tree, concept provenance, and sampling metadata
- keep artifact generation reproducible and machine-readable

Reason:

- synthetic data is now a first-class research instrument rather than a side utility

### 36. Equivalence Family Corpus Format

Define a corpus format for multiple trees that are related to the same law.

Targets:

- record equivalence families containing:
  - one anchor tree
  - one or more alternative trees
  - the declared relation type
- support at minimum:
  - exact same raw tree
  - normalized same raw tree
  - known concept-level equivalence
  - sampled numeric equivalence
  - contextual subtree substitution
- keep relation types explicit rather than collapsing them into one label

Reason:

- the project now needs to study EML combinatorics, not just single-tree recovery

### 37. Paired And Grouped Dataset Generator

Generate datasets tied to equivalence families rather than only one originating tree.

Targets:

- produce paired or grouped datasets where multiple trees yield the same sampled outputs
- preserve which trees belong to the same family
- support multiple sampling domains for the same family
- emit artifacts that the future ML layer can use directly

Reason:

- the model should be exposed deliberately to cases where different tree forms fit the same data

### 38. Snippet-Level Dataset Generator

Create corpora aimed at partial-law discovery rather than only whole-formula recovery.

Targets:

- generate larger target trees with labeled subtrees
- emit datasets for whole targets and for selected partial subregions
- preserve snippet provenance and parent-tree context
- support overlap between snippets so later assembly tasks are possible

Reason:

- the intended ML behavior is to recover useful subtrees and partial laws, even when no full global law is found

### 39. Partial-Fit Evaluation Methodology

Extend the experiment methodology beyond top-1 whole-formula recovery.

Targets:

- define what counts as a useful partial recovery
- distinguish:
  - full-law recovery
  - snippet recovery
  - equivalence-family recovery
  - compatible but incomplete assembly
- define evaluation criteria for partial explanations over data subsets or regions
- document what claims partial recovery can support in later empirical writing

Reason:

- the project now values partial laws as first-class outputs rather than mere failures

### 40. Partial-Fit Experiment Harness

Build a harness around snippet-level and equivalence-family experiments.

Targets:

- load paired, grouped, and snippet-level experiment specs
- generate or load their datasets and equivalence-family artifacts
- run baseline matching or ranking workflows
- record artifact-level outputs suitable for later ML comparison
- keep the first harness deterministic and reproducible

Reason:

- the new direction needs an experiment apparatus, not just an idea in docs

### 41. Equivalence-Aware ML Baseline

Implement the first Python-side baseline over Go-generated corpora.

Targets:

- train or evaluate a simple baseline that maps data windows or sample sets to candidate snippets
- expose it to families where multiple trees yield the same outputs
- measure whether it learns interchangeable subtree behavior better than a single-target baseline
- keep the first baseline intentionally simple and inspectable

Reason:

- the value of the pivot must be tested empirically, not only argued architecturally

### 42. Assembly-Oriented Evaluation Suite

Evaluate whether recovered snippets can participate in larger coherent constructions.

Targets:

- define assembly-style experiments over snippet candidates
- measure compatibility between recovered pieces
- preserve useful partial outputs when full assembly fails
- compare full-recovery and partial-recovery usefulness explicitly

Reason:

- the intended payoff is not only exact end-to-end recovery, but useful partial structure discovery that humans can build on

### 43. Search Repositioning

Reposition search as one subsystem inside the broader equivalence-learning program.

Targets:

- document how the existing search track fits into the new monorepo direction
- identify which search improvements still support snippet discovery and equivalence studies
- separate search work that is still strategically useful from search work that can remain deferred
- keep oracle whole-formula search as a baseline rather than the sole success criterion

Reason:

- search still matters, but it should now serve the broader combinatorics and ML agenda rather than dominate it

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
