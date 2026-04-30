# Search

This directory holds Go-side search algorithms over raw EML trees.

Current subpackages:

- `common/`
  - shared candidate, target, scoring, retention, bounds, and tree helpers
- `enumerative/`
  - bounded enumerative and layered baseline search
- `maze/`
  - single-threaded frontier-aware growth-thread search for partial-law discovery
- `lookup/`
  - reserved for future fast signature/database matching
- `mutate/`
  - reserved for future local refinement and replacement-driven search

The top-level `search` package remains the stable facade for existing callers.
New work should prefer the subpackages for algorithm-specific code.

Search families should own traversal and candidate generation, not repository
identity. Scoring, target interpretation, and retain/prune semantics should be
provided through shared adapters so full-match, partial-match, and ML-guided
routes can evolve in parallel.

Shared adapters now cover both full-match scoring and first-pass
partial-coverage scoring.
