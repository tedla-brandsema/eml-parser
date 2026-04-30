# ML Subproject

This directory is reserved for the future Python ML subproject.

The Python side of the repository is expected to:

- consume generated corpora from `artifacts/`,
- consume experiment artifacts from `experiments/` when useful,
- train and evaluate models for snippet discovery, equivalence-aware ranking,
  and assembly experiments,
- avoid reimplementing symbolic semantics that already belong to the Go core.

Item 33 only reserves this location and records the repository contract. It
does not yet define:

- a Python package layout,
- a dependency manager,
- a training framework,
- or the first executable ML experiment.

Those choices are deferred to item 34 in `TODO.md`.
