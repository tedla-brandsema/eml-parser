# `search/maze`

Frontier-aware growth-thread search over partial EML trees.

V1 is intentionally:

- Go-only
- single-threaded
- deterministic
- seeded from explicit anchors
- limited to atomic growth moves
- scored against the whole dataset in this phase

The purpose is not to find a single perfect formula. The purpose is to:

- grow partial EML shapes outward from anchors,
- grow from explicit frontier locations inside the current tree,
- validate each growth step against data,
- retain multiple surviving branches,
- prune or retreat when branches stop matching,
- and preserve dead-end partials as useful outputs.

Future optimization may evaluate multiple branches concurrently, but concurrency
is not part of this phase. Partial-dataset coverage scoring is also deferred to
the next maze step, after frontier growth is stable.

Current whole-dataset fit and threshold behavior are default scorer and
retention-policy choices, not the permanent identity of maze search.

Maze can now also be seeded from curated snippet artifacts as an explicit
anchor source. Automatic data-to-snippet matching is still deferred.
