# `search/maze`

Naive growth-thread search over partial EML trees.

V1 is intentionally:

- Go-only
- single-threaded
- deterministic
- seeded from explicit anchors
- limited to atomic growth moves

The purpose is not to find a single perfect formula. The purpose is to:

- grow partial EML shapes outward from anchors,
- validate each growth step against data,
- retain multiple surviving branches,
- prune or retreat when branches stop matching,
- and preserve dead-end partials as useful outputs.

Future optimization may evaluate multiple branches concurrently, but concurrency
is not part of the first implementation.
