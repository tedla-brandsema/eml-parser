# Search

This directory holds Go-side search algorithms over raw EML trees.

Current subpackages:

- `common/`
  - shared candidate, scoring, bounds, and tree helpers
- `enumerative/`
  - bounded enumerative and layered baseline search
- `maze/`
  - single-threaded growth-thread search for partial-law discovery
- `lookup/`
  - reserved for future fast signature/database matching
- `mutate/`
  - reserved for future local refinement and replacement-driven search

The top-level `search` package remains the stable facade for existing callers.
New work should prefer the subpackages for algorithm-specific code.
