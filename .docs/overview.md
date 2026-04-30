# EML Parser Overview

## Summary

`eml-parser` starts as a Go project for parsing and representing EML expressions under full project control. The parser remains intentionally narrow: it only accepts the atomic EML language from the paper. Mathematical richness is introduced above that layer through a recursive dictionary of named concepts that expand down to raw EML trees.

The project should now be understood as a monorepo in spirit:

- the Go portion is the deterministic symbolic core,
- a future Python `ml/` subproject is the statistical learning and assembly layer,
- and shared experiment artifacts connect the two.

The broader motivation comes from two linked ideas:

- EML offers a uniform binary-tree representation for elementary functions.
- Code-oriented AI systems work best when they can generate structures, run verifiers, inspect failures, and repair them iteratively.

Together, those suggest a workflow where mathematical expressions can be treated more like programs: expanded from reusable concept blocks, evaluated, searched, transformed, grouped into equivalence families, and eventually verified.

## Why EML Matters

The EML paper argues that a single binary operator plus a distinguished constant can express the ordinary scientific-calculator repertoire of elementary functions. The important engineering consequence is not only expressive completeness, but structural regularity.

Instead of a heterogeneous grammar with many unrelated operators, EML expressions can be represented as binary trees of one operator applied repeatedly to terminals and intermediate results. That makes EML a plausible substrate for:

- symbolic regression over a complete expression family,
- equivalence-aware snippet discovery over large datasets,
- compositional assembly from partial laws,
- compiler-style normalization and rewriting,
- interchangeable evaluation backends,
- eventual export into proof-oriented systems such as Lean.

This project treats EML first as a language and IR problem: define the atomic syntax, parse it reliably, own the AST, and keep later concept expansion, search, and proof features possible.

## Long-Term Vision

The larger system this parser could support is a code-oriented mathematical pipeline:

1. infer candidate EML snippets from data or search,
2. represent them in a uniform EML tree form,
3. compare them inside equivalence neighborhoods rather than as single isolated trees,
4. attempt larger assemblies from compatible partial laws,
5. evaluate them with increasingly strict numeric backends,
6. translate them into a formal environment for proof or verification.

In that model, symbolic regression is closer to constrained program synthesis than to ad hoc formula guessing. The important unit is not always a single globally correct tree. It may be a family of interchangeable local trees that explain part of a dataset and can later be assembled into larger explanations. Lean or another formal system then becomes the equivalent of a type checker or proof verifier. The raw parser and AST are the foundation for that workflow, but they are not the place where higher-level mathematical concepts live.

## Current Project Priority

The project currently has two possible outcomes:

- it may become an empirical research vehicle for evaluating EML-based symbolic-regression workflows,
- and, if those experiments are promising, it may later harden into a reusable third-party tool.

The order matters.

For now, the project should be treated primarily as an experimental apparatus for testing whether this EML-centered approach is actually useful. Broader tool hardening, polish, and third-party usability are conditional on positive experimental evidence rather than assumed in advance.

That means current work should prioritize:

- experiment validity,
- reproducibility,
- interpretability of equivalence behavior and search behavior,
- and credible oracle-controlled evaluation.

General-purpose tool polish is still useful, but it is downstream of proving that the direction is worth pursuing.

The immediate research pivot is:

- synthetic datasets generated from known raw EML trees and known concept expansions,
- paired or grouped EML trees that produce the same outputs,
- partial-law discovery rather than only whole-formula recovery,
- and a two-stage workflow:
  - data to candidate trees or snippets,
  - candidate trees or snippets to equivalence classes and larger assemblies.

## Immediate Goal

The first concrete target for `eml-parser` is a parser foundation in Go:

- lexer,
- parser,
- AST,
- evaluator interfaces,
- clear semantics around precision and branch behavior.

This is intentionally narrower than "build symbolic regression" or "integrate Lean". Those remain future consumers of the same internal representation.

The next substantial layer above the Go core is no longer just "better search". It is a monorepo split:

- Go for parsing, symbolic structures, equivalence-family generation, and synthetic corpus generation,
- Python for ML experiments over those corpora, especially snippet discovery, equivalence-aware learning, and assembly experiments.

## Architectural Split

The project is intentionally divided into two layers.

### Raw EML Layer

This is the layer owned by the parser, AST, and evaluator.

- It only knows the atomic EML language:
  - the distinguished constant `1`,
  - variables,
  - binary `eml(left, right)` application.
- No matter how large a final expression becomes, it is still just a composite of these atomic forms.
- The parser should stay minimal for this reason. It does not need to understand named mathematical concepts such as `sin`, `cos`, or `tan`.

### Concept Dictionary Layer

This is the layer where mathematical meaning is organized.

- Named concepts are stored as reusable mappings.
- A concept may expand directly to raw EML, or it may be defined in terms of lower-level concepts.
- Expansion continues recursively until only raw EML remains.
- The result of expanding any concept is still a raw EML tree, which means it can be parsed, evaluated, composed, normalized, and stored like any other raw expression.

This should be thought of as a Euclid-like construction hierarchy:

- small atomic basis,
- named derived constructions,
- each derived construction reducible to earlier constructions,
- full reduction ending in atomic EML only.

Conceptually, this means examples like `tan` do not need to be parsed as language syntax. Instead:

- `tan` is a dictionary key,
- its body may refer to lower-level concept keys such as `sin` and `cos`,
- those in turn reduce further,
- `tan.EML` means "fully expand the concept until only raw EML remains."

This is the key design constraint for the project: keep the parser small, and move mathematical richness into the concept dictionary plus its expansion engine.

## Technical Direction

### Parser and AST

The parser should be built in Go with a grammar and AST fully controlled by the project. `goyacc` is the baseline parsing approach because it gives explicit control over tokens, productions, AST construction, and error handling.

The raw EML AST should be designed to support more than evaluation. It should be suitable for:

- pretty-printing,
- serialization,
- tree rewrites,
- normalization,
- search over structure,
- later export to proof or CAS tooling.

The core language shape is:

- constants,
- variables,
- binary `eml(left, right)` application.

The parser boundary should expose:

- position-aware lexer tokens,
- parse errors with source locations,
- a parse entrypoint that returns a typed raw EML AST.

### Concept Dictionary and Expansion

The concept layer should be implemented separately from the parser.

At minimum, it should provide:

- a registry of named concepts,
- parameterized concept definitions,
- concept bodies that can reference lower-level concepts,
- recursive expansion from concept expressions to raw EML ASTs,
- cycle detection and unknown-concept errors,
- arity checking on concept calls.

The expansion engine is the bridge between the mathematical dictionary and the parser/evaluator layer:

- concept expressions are human-manageable reusable blocks,
- expanded expressions are raw EML trees suitable for execution and storage.

### Monorepo Split

The project should now evolve as a monorepo with a narrow interface between symbolic generation and ML experimentation.

The Go side should own:

- raw EML parsing and AST,
- concept expansion,
- evaluation backends,
- normalization used only for hygiene and controlled comparison,
- synthetic dataset generation,
- equivalence-family generation,
- experiment artifact generation.

The Python `ml/` side should own:

- model training,
- snippet ranking or generation,
- equivalence-aware learning objectives,
- assembly experiments over partial trees,
- evaluation against shared experiment artifacts generated by Go.

The interface between them should be artifact-driven rather than ad hoc:

- JSON specs,
- generated datasets,
- equivalence-family corpora,
- result artifacts.

### Evaluation Backends

Evaluation should not be hard-wired to one numeric representation. The project should define an evaluator interface that allows multiple backends with identical raw-AST traversal and explicit semantics.

At minimum, the architecture should support:

- a fast `float64` or `complex128` backend for broad screening,
- a higher-precision backend for validation and identity checking,
- explicit treatment of principal-branch semantics and exceptional values.

## Precision Strategy

`float64` is useful, but not sufficient as the authoritative numeric substrate.

Reasons:

- repeated `exp` and `log` compositions can overflow or underflow rapidly,
- branch cuts and cancellation can make near-equalities misleading,
- symbolic-regression or equivalence workflows need stronger separation between approximate agreement and real identity,
- complex arithmetic amplifies representational and domain issues.

The design assumption for this project is:

- use native floating-point for speed,
- use higher precision for anything trusted,
- make precision a configurable policy rather than an accidental implementation detail.

One consequence is that approximate numeric agreement is not identity. This matters even more once the project studies families of different trees that fit the same sampled data.

Go's `math/big` package covers arbitrary-precision real arithmetic, but not a complete high-precision transcendental stack. That means the evaluator layer must stay pluggable. The parser and AST should not assume a specific final numeric engine.

## Non-Goals for V1

Version 1 of this project does not aim to deliver:

- a full symbolic regression engine,
- Lean integration,
- theorem proving features,
- a final choice of arbitrary-precision transcendental library,
- a complete optimizer or search framework,
- a finished ML training stack.

V1 is about building the language foundation cleanly enough that these remain possible.

## Initial Milestones

The next implementation steps are:

1. Keep the raw EML grammar minimal and stable.
2. Maintain the raw AST, parser, and evaluator as the execution substrate.
3. Add a concept dictionary with parameterized definitions and recursive expansion.
4. Define a small initial standard library of grounded concept mappings.
5. Add normalization and reuse around expanded raw EML trees.
6. Continue improving high-precision evaluation without coupling the parser to any richer surface syntax.
7. Generate synthetic datasets and equivalence families from the Go core.
8. Feed those artifacts into a future Python ML layer aimed at snippet discovery and compositional assembly.

Current status:

- the parser foundation exists,
- the parser is generated from `parser/eml.y` using `goyacc`,
- a first evaluation layer should treat `eml(a, b)` as `exp(a) - log(b)` in the complex domain using principal-branch semantics, matching the paper's EML reading.
- the parser intentionally supports only atomic EML, not named mathematical concepts,
- the fast backend uses `complex128`,
- the higher-precision path is defined as a boundary, not yet a concrete backend: explicit precision metadata, explicit logarithm-branch semantics, and an opaque high-precision value type.
- a pure-Go high-precision backend now exists around `big.Float`-backed complex values using `github.com/mshafiee/bigmath` for the real transcendental layer,
- principal-branch complex `exp` and `log` are implemented from real `exp`, `log`, `sin`, `cos`, and `atan2`,
- unsupported logarithm branches remain explicit errors rather than silent behavior changes.
- PoC-style tests cover nested expression trees, end-to-end parser/evaluator behavior, and agreement between the fast and high-precision backends on representative expressions.
- the next architectural layer is a concept dictionary that expands named mathematical blocks into raw EML trees without broadening the parser.
- the concept dictionary now includes a first reusable standard library spanning constants, arithmetic, `sqrt`, hyperbolic functions, and basic trigonometric functions, all defined compositionally and expanded to raw EML on demand.
- tooling should now focus on introspecting that dictionary: listing concepts, showing definitions, tracing dependencies, and emitting symbolic raw-EML expansions without changing the raw parser.
- raw EML normalization should now reduce obviously redundant expanded forms after concept expansion, without changing the concept dictionary or parser grammar.
- normalization should now be treated as a controlled tool for hygiene and comparison, not as the only definition of identity.
- expansion caching should now avoid repeated recursive work for named concepts while staying internal to the registry layer.
- search-space preparation should now remain available as one experiment track, but it is no longer the only or highest-priority path.
- dataset and benchmark support should now expand toward synthetic equivalence corpora, paired-tree datasets, and snippet-level training artifacts rather than only whole-formula oracle search.
- the first search skeleton and diagnostics remain useful baselines, but they should now be treated as one consumer of the symbolic core rather than the project's primary destination.
- the formalization bridge should now export normalized raw EML into a deterministic proof-friendly intermediate form with retained concept provenance where available, without coupling proof concerns back into the parser or concept registry.
- oracle-controlled experiment methodology should now be fixed in `.docs/experiments.md` before larger empirical result sets are generated, so recovery classes and publishable claims are defined in advance rather than inferred later.
- the experiment filesystem layout should now be fixed under `experiments/` so specs, datasets, results, and suite reports have predictable locations and naming before repeated runs begin producing artifacts.
- the experiment schema should now be fixed in code and in example JSON specs so future runners load declarative oracle experiments rather than embedding them implicitly in tests or ad hoc scripts.
- the oracle dataset generator should now be able to materialize deterministic real-valued datasets from concept or raw targets into `experiments/datasets/`, using the current evaluation stack and failing loudly on non-real outputs.
- the oracle search harness should now be able to load one experiment spec, generate or load its dataset, run the current enumerative real search, and persist a machine-readable result artifact in `experiments/results/`.
- experiment runs should now be classified deterministically as exact normalized recovery, concept-equivalent recovery, approximate-only recovery, or no recovery according to the declared recovery criteria in each spec.
- experiment result artifacts should now retain enough provenance to be auditable later: dataset metadata, target canonical form, search options, diagnostics, ranked candidates, recovery class, timestamp, and git commit when available.
- suite reporting should now be able to aggregate explicit result artifacts into one JSON summary and one Markdown summary under `experiments/reports/`, including recovery-class counts, target-family counts, diagnostic ranges, and top recovered expressions.
- an initial oracle suite should now exist as committed experiment specs covering exact controls, a small nested composite control, and honest current-boundary failures such as additive and larger-library targets.
- paper-readiness notes should now make explicit what the current suite can support as evidence, what it cannot yet support, and which threats to validity must be disclosed if the project is written up.
- the next research layer should now focus on synthetic data plus EML combinatorics:
  - multiple trees for the same law,
  - local subtree substitutions,
  - partial-law recovery,
  - and later ML-based assembly over snippets rather than only top-1 global recovery.

## Defaults and Assumptions

- `.docs/overview.md` is the initial source of truth instead of a larger doc tree.
- The raw parser should remain minimal even as the mathematical concept layer grows.
- Mathematical concepts belong in a dictionary/expansion layer, not in the raw grammar.
- `goyacc` is the default parser strategy unless later experience shows a simpler parser architecture is clearly better.
- Precision and branch semantics are first-class design concerns, not secondary implementation details.
