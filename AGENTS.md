# AGENTS

## Purpose

This file records the stable architectural constraints for work in this repository.

Use `TODO.md` for implementation sequence and backlog.
Use this file for the rules that should continue to hold as the project grows.

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

## Long-Term Direction

The long-term objective is not only evaluation, but generation and discovery of EML trees from data.

That means future work should continue to preserve:

- a minimal raw substrate,
- a reusable concept dictionary,
- expansion to executable raw trees,
- measurable and normalizable raw expressions,
- a clean path toward symbolic regression and later proof-oriented workflows.
