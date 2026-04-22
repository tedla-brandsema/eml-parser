# Concept Dictionary

## Purpose

The project has two different responsibilities that must remain separate.

- The raw parser owns the atomic EML language.
- The concept dictionary owns named mathematical constructions.

This document fixes that split so the implementation does not drift.

## Raw Parser

The raw parser must remain minimal.

It only parses:

- `1`
- variables
- `eml(left, right)`

That is sufficient because every final expression, no matter how large, is still an EML tree composed of these atomics.

The parser is not where `sin`, `cos`, `tan`, or other named concepts belong.

## Concept Dictionary

Named mathematical concepts live in a separate registry.

- Each concept has a name.
- Each concept may have parameters.
- Each concept body may be:
  - raw EML composition,
  - references to lower-level concepts,
  - or a mixture of both.

Examples conceptually:

- `exp(x)` may reduce directly to a raw EML tree.
- `tan(x)` may reduce to lower-level concepts such as `sin(x)` and `cos(x)`.
- those lower-level concepts may reduce further until only raw EML remains.

The result of full expansion is always a raw EML tree.

## Expansion Rule

Expansion is recursive.

1. Start from a named concept and its arguments.
2. Replace concept references with their definitions.
3. Continue until no concept references remain.
4. The final result is a raw EML AST composed only of:
   - `1`
   - variables
   - `eml(left, right)`

That expanded raw EML tree is then the executable form used by the evaluator and any later normalization or search logic.

## Design Constraints

- Do not broaden the raw parser just to support named concepts.
- Do not mix concept definitions into the raw EML grammar.
- Do keep concept definitions composable and recursively expandable.
- Do support partial trees that combine raw EML nodes with concept references.
- Do detect missing concepts, wrong arity, and cycles during expansion.

## Current Direction

The first implementation should provide:

- a concept registry,
- parameterized concept definitions,
- concept references inside concept bodies,
- recursive expansion to raw EML AST,
- a small grounded standard library that only contains concepts we can justify from the current implementation.

The current standard-library growth strategy is conservative:

- start from paper-grounded `exp`,
- add direct EML constants or identities that follow from the paper's `eml(a, b) = exp(a) - log(b)` semantics,
- define additional arithmetic concepts compositionally in terms of already-grounded concepts,
- prefer correctness and recursive composability first, and shortest-known EML trees later.

This means early standard-library entries may not be the shortest possible EML witnesses from the paper's discovery chain, but they must still expand fully to valid raw EML trees.

The current library now covers:

- foundational constants: `one`, `zero`, `e`, `minus_one`, `two`, `half`, `i`, `pi`
- direct exp-log layer: `exp`, `log`, `id`
- arithmetic layer: `sub`, `neg`, `add`, `recip`, `mul`, `div`, `pow`, `square`, `sqrt`
- hyperbolic layer: `sinh`, `cosh`, `tanh`
- trigonometric layer: `sin`, `cos`, `tan`

These are still concept-layer definitions, not parser syntax. They become executable only after recursive expansion to raw EML.

## Tooling

Tooling should operate on the concept layer, not by changing the parser.

The first registry tooling surface should provide:

- listing registered concept names,
- showing a concept definition in concept-layer form,
- showing direct and transitive dependencies,
- expanding a named concept symbolically into raw EML by binding parameters to variables of the same name.

This makes the concept library inspectable and debuggable while preserving the architectural split:

- parser for raw EML only,
- dictionary and tooling for mathematical concepts.

Normalization remains a raw-AST concern after expansion.

- expand concepts to raw EML first,
- then normalize raw EML,
- do not normalize by mutating concept definitions.

Expansion caching should stay behind the registry boundary.

- cache expanded raw EML trees for named concepts,
- keep cache invalidation tied to concept registration changes,
- do not push caching concerns into the parser or evaluator APIs.
