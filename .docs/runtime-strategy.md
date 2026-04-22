# Runtime Strategy

## Purpose

This document records the current runtime choice and the conditions under which
the project may eventually move to a more native systems-oriented stack such as
Zig.

It is not a commitment to migrate now. It is a guide for keeping today's design
portable enough that a migration remains realistic later.

## Current Choice: Go

The project currently uses Go because it already supports the core structure we
need:

- a narrow raw EML parser,
- a recursive concept dictionary,
- expansion to raw EML AST,
- interchangeable evaluation backends,
- normalization and search-preparation tooling,
- fast iteration on architecture and validation.

Go is good enough for the current stage because the main bottlenecks are
semantic and mathematical, not yet systems-level.

## Why Zig Is A Plausible Future Target

Zig becomes attractive if the project needs a stronger native runtime layer for:

- arbitrary-precision numeric backends based on mature C libraries,
- lower-overhead interop with MPFR, MPC, GMP, or similar libraries,
- tighter control over memory layout and allocation in search-heavy workflows,
- easier packaging of a native parser/runtime stack around C ecosystem tools.

Zig is especially plausible for the numeric core and search runtime, not
because Go is structurally wrong, but because native-library interop and
lower-level control may eventually dominate the tradeoffs.

## Important Caveat

The current `parser/eml.y` is a `goyacc` grammar, not a drop-in C yacc grammar.

That means:

- the grammar productions are conceptually reusable,
- but the semantic actions are written in Go and would need to be rewritten for
  a C or Zig parser pipeline,
- so migration would be a controlled port, not a direct generator swap.

## What Should Remain Portable

To keep a future migration realistic, the project should continue to preserve
these boundaries:

### Raw EML Model

- keep the raw AST small and language-neutral:
  - `1`
  - variables
  - `eml(left, right)`

This is the best migration boundary in the project.

### Concept Dictionary

- keep concept definitions declarative and recursive,
- avoid tying concept semantics to Go-specific runtime assumptions,
- treat concept expansion as a portable transformation from named definitions to
  raw EML trees.

### Normalization Rules

- keep normalization expressed as raw-tree rewrite logic,
- avoid making normalization depend on Go reflection or other language-specific
  conveniences,
- keep the rewrite rules clear enough that they can be ported.

### Evaluation Interfaces

- keep backend interfaces narrow,
- preserve explicit precision and branch semantics,
- avoid leaking implementation details from `bigmath` into the broader project
  model.

The evaluator is the most likely subsystem to be replaced first.

### Search Utilities

- keep search preparation focused on raw EML trees and normalized canonical
  forms,
- avoid coupling search utilities to concept registration internals,
- treat candidate scoring and tree utilities as portable algorithmic layers.

## What Would Likely Change First In A Migration

If the project eventually moves toward Zig, the likely order is:

1. high-precision numeric backend,
2. performance-sensitive search/runtime code,
3. parser/runtime integration,
4. only then any broader language-level migration.

That means the most realistic early migration is a native backend swap, not a
full rewrite of the project.

## Migration Triggers

The project should consider a serious Zig migration only if one or more of
these become persistent blockers:

- the Go high-precision backend remains the main semantic limitation,
- native numeric libraries become necessary for correctness rather than just
  optimization,
- cgo-style integration in Go becomes too awkward or fragile,
- search/runtime performance becomes dominated by allocation or interop costs,
- packaging a native parser/runtime stack becomes more valuable than staying in
  pure Go.

Absent these triggers, staying in Go is the default.

## Practical Rule For Current Work

Do not optimize today's implementation around an immediate migration.

Instead:

- keep interfaces narrow,
- keep core representations portable,
- keep semantics explicit,
- avoid unnecessary Go-specific coupling.

That preserves optionality without slowing down the current project.
