# Artifacts

This directory is reserved for shared generated corpora that sit between the Go
symbolic core and the future Python `ml/` subproject.

Unlike `experiments/`, which is specific to oracle experiment specs and their
run outputs, `artifacts/` is the shared corpus area for generated data that may
be consumed by multiple downstream workflows.

Current reserved subdirectories:

- `artifacts/equivalence/`
  - generated equivalence-family corpora
- `artifacts/snippets/`
  - generated snippet and partial-law corpora

Generated artifacts here are reproducible and ignored by default.

The source of truth for ownership and read/write direction is
`.docs/artifact-contract.md`.
