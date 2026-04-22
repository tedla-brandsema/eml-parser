# BACKLOG

This file records unresolved issues, known limitations, and explicit workarounds
accepted during implementation.

Use this file when:

- a feature is only partially implemented,
- a test or validation had to be narrowed to match current behavior,
- a semantic mismatch is known but not yet fixed,
- a backend limitation blocks the ideal implementation,
- a workaround was chosen to keep progress moving.

Do not silently pass these by. If we discover one, we add it here.

## Open Items

### 1. Inverse Function Branch Semantics

Status: open

During standard-library validation, some inverse-function branch-sensitive cases
did not match Go's `math/cmplx` principal-branch results exactly.

Observed behavior:

- `asin(2)` returned the opposite sign on the imaginary component relative to
  `cmplx.Asin(2)`.
- `atanh(2)` returned the opposite sign on the imaginary component relative to
  `cmplx.Atanh(2)`.

What we did:

- narrowed branch-sensitive validation to the cases the current implementation
  matches reliably: `log`, `sqrt`, and `acosh`.
- kept the inverse-function concepts in the standard library because they work
  for the non-branch-cut reference cases currently tested.

Why this matters:

- strict principal-branch alignment is required for trustworthy inverse
  function behavior across the full complex plane.

Long-term follow-up:

- determine whether the current concept definitions should be replaced with
  alternative formulas,
- or whether the evaluator / normalization stack needs explicit branch-aware
  handling for these constructions.

### 2. High-Precision Backend Exact-Zero Edge Cases

Status: open

The current `bigmath`-backed high-precision backend still fails on some
compositional concept evaluations with:

- `logarithm undefined for zero`

Observed during validation attempts for more complex formulas such as:

- some inverse-function validation paths,
- some `sqrt` / nested arithmetic compositions,
- some deeper compositions built from current concept-level arithmetic.

What we did:

- restricted representative high-precision validation to concepts that the
  current backend supports reliably: `mul`, `sigmoid`, and `log`.
- avoided claiming that all concept-level formulas are currently safe under the
  high-precision backend.

Why this matters:

- symbolic-regression and trusted evaluation workflows will eventually depend on
  the high-precision path handling intermediate values robustly.

Long-term follow-up:

- harden the high-precision backend against intermediate zero / branch-cut
  cases,
- review arithmetic concept definitions that create fragile intermediate forms,
- or introduce better high-precision semantics for exceptional values.

### 3. Validation Thresholds Are Empirical, Not Proven

Status: open

Some normalization and expansion-size checks were initially guessed too
aggressively and had to be adjusted to match the actual current implementation.

Examples:

- normalized depth thresholds for `sigmoid` and `atanh`,
- expansion-size lower bounds for `atanh` and `sigmoid`.

What we did:

- changed these checks to realistic thresholds based on current expanded and
  normalized trees.

Why this matters:

- the current thresholds are useful regression guards, but they are empirical
  baselines, not formal complexity guarantees.

Long-term follow-up:

- replace ad hoc thresholds with more principled expectations,
- possibly tie them to canonicalized forms or tracked complexity snapshots.

### 4. Standard Library Validation Coverage Is Intentionally Incomplete

Status: open

The current validation layer is stronger than before, but it is not exhaustive.

Known gaps:

- no full principal-branch validation for all inverse trigonometric functions,
- no broad high-precision coverage for the whole standard library,
- no formal proof that concept definitions preserve intended semantics.

What we did:

- implemented strong `complex128` reference checks,
- added branch-sensitive checks for the cases that currently behave reliably,
- added representative high-precision checks where the backend is stable.

Why this matters:

- future users could otherwise misread the current test suite as a proof that
  all branches and all backends are fully aligned.

Long-term follow-up:

- continue expanding validation incrementally,
- keep recording any narrowed or skipped validation here.
