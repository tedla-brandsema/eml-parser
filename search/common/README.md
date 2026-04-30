# `search/common`

Shared helpers for Go-side search algorithms.

This package re-exports the stable raw-tree candidate, target, scoring,
retention, bounds, and tree-inspection utilities used by multiple search
strategies.

Algorithms should consume these shared policies rather than hard-code a single
search objective into their own control flow.

The shared adapter layer now supports both whole-dataset scorers and
coverage-aware local-window scorers.
