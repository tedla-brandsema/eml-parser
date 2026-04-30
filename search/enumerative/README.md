# `search/enumerative`

Bounded baseline search algorithms.

This package exposes the existing enumerative and layered real-valued search
strategies behind a dedicated algorithm namespace while the top-level `search`
package remains stable for current callers.

Current exact or whole-dataset behavior is a default scorer/policy combination,
not the permanent definition of enumerative search.
