# AGENTS

## Purpose

This file records the stable architectural constraints for work in this repository.

Use `TODO.md` for implementation sequence and backlog.
Use this file for the rules that should continue to hold as the project grows.

## Backlog Discipline

Use `BACKLOG.md` for unresolved issues, known limitations, semantic mismatches,
and explicit workarounds.

If implementation requires narrowing validation, accepting a workaround, or
skipping an intended behavior because of a known limitation, record it in
`BACKLOG.md`.

Do not silently pass these by.

## Commit Messages

Always use Conventional Commits for git commit messages.

Examples:

- `feat: add inverse trig concepts`
- `fix: normalize exp(log(x)) to x`
- `docs: update concept dictionary notes`

## Core Split

The project is intentionally split into separate layers:

- raw EML substrate
- concept dictionary
- expansion
- evaluation
- tooling

The project also now has a monorepo-level split:

- Go is the deterministic symbolic core
- Python `ml/` will be the statistical learning and assembly layer
- shared artifacts are the contract between them

Do not collapse these layers together casually.

## Raw Parser Rules

The raw parser stays minimal.

It only supports:

- `1`
- variables
- `eml(left, right)`

Do not broaden the raw parser just to support named mathematical concepts.

Named concepts such as `exp`, `log`, `sin`, `cos`, `tan`, and future functions do not belong in the raw grammar.

## Concept Dictionary Rules

Named mathematical concepts belong in the concept dictionary layer.

- Concepts may be parameterized.
- Concept bodies may mix raw EML composition with references to lower-level concepts.
- Concepts must be recursively expandable until only raw EML remains.
- Expanded concepts must reduce to raw EML AST composed only of:
  - `1`
  - variables
  - `eml(left, right)`

Do support:

- partial trees that combine raw EML nodes with concept references
- recursive composition of named concepts
- cycle detection
- missing-concept errors
- arity checking

Do not mix concept definitions into the parser grammar.

## Expansion Rules

Expansion is the bridge from concept-layer mathematics to executable raw EML.

- `show`-style tooling should expose concept-layer definitions.
- `expand`-style tooling should expose fully expanded raw EML.
- Expansion should preserve the parser/evaluator contract by returning raw EML AST.

If a concept is inspectable or executable, it should be possible to reduce it to raw EML without changing parser behavior.

## Evaluation Rules

Evaluation operates on raw EML AST, not on concept-layer expressions.

- Keep backends interchangeable.
- Preserve explicit precision policy.
- Preserve explicit logarithm-branch semantics.
- Avoid hiding branch-sensitive behavior behind silent defaults.

The fast backend is for screening.
Higher-precision backends are for more trusted evaluation.

## Tooling Rules

Tooling should inspect the concept dictionary and expanded raw EML without broadening the raw language.

Useful tooling includes:

- list concepts
- show concept-layer definitions
- trace direct and transitive dependencies
- expand concepts symbolically to raw EML
- measure expanded-tree size and depth

Tooling should help quantify and inspect the dictionary, not replace the separation between dictionary and parser.

## Standard Library Growth Rules

Grow the concept standard library conservatively.

- Prefer grounded, compositional definitions.
- Prefer correctness and inspectability over shortest-known constructions.
- Keep additions in the concept layer unless the raw EML substrate itself truly changes.
- Extend the dictionary without changing the parser unless there is a compelling substrate-level reason.

Current library categories include:

- foundational constants
- exp/log layer
- arithmetic
- hyperbolic functions
- trigonometric functions

## Normalization Rules

Normalization belongs after expansion.

- Normalize raw EML AST, not concept definitions.
- Keep normalization separate from parsing.
- Keep normalization separate from concept registration.
- Use normalization for hygiene, controlled comparison, and conservative dedupe.
- Do not treat normalization as the only definition of identity.
- Preserve room for later equivalence analysis between structurally different trees.

## Search And Equivalence Rules

Search is now only one consumer of the symbolic core.

- Do not assume top-1 exact tree recovery is the only meaningful output.
- Prefer work that supports synthetic datasets, equivalence families, snippet discovery, and partial assembly.
- Treat equivalence analysis as a distinct stage from raw candidate generation.
- Avoid premature normalization that would erase interesting combinatoric structure before it can be studied.
- Search algorithms must not hard-code a single discovery objective as repository identity.
- Scoring, target interpretation, and retain/prune semantics must remain adapter-based so full-match, partial-match, and ML-guided discovery can coexist.
- New search work must preserve the ability to pursue:
  - exact full recovery
  - partial-law discovery
  - ML-guided ranking or seeding
- If an implementation temporarily narrows one discovery route, record that narrowing in `BACKLOG.md`.

## Monorepo Rules

The Go side should own:

- symbolic semantics,
- parser and AST,
- concept expansion,
- evaluation,
- synthetic data generation,
- equivalence-family generation,
- experiment artifacts.

The Python `ml/` side should own:

- model training,
- snippet ranking or generation,
- equivalence-aware learning,
- assembly experiments.

Do not duplicate symbolic semantics in Python when the Go core can generate the needed artifact deterministically.

Directory ownership should stay explicit:

- Go code remains at repo root
- `experiments/` is for oracle experiment inputs and outputs
- `artifacts/` is for shared Go-generated corpora consumed by `ml/`
- `ml/` is for Python-side learning and assembly work

## Long-Term Direction

The long-term objective is not only evaluation, but generation and discovery of EML trees from data.

That means future work should continue to preserve:

- a minimal raw substrate,
- a reusable concept dictionary,
- expansion to executable raw trees,
- measurable and normalizable raw expressions,
- a clean path toward symbolic regression and later proof-oriented workflows,
- and a clean path toward equivalence-aware ML over synthetic corpora and partial laws.
