# eml-parser

A Go implementation and empirical research vehicle for the **EML** (Exp-Minus-Log) operator, based on the paper [All elementary functions from a single operator](https://arxiv.org/html/2603.21852v2).

---

## What Is EML?

EML defines a single binary operator:

```
eml(x, y) = exp(x) − ln(y)
```

paired with the single constant `1`. The paper's central claim — analogous to the universality of the NAND gate in boolean logic — is that this minimal two-symbol system can express every elementary function a scientific calculator provides. For example:

| Function | EML expression |
|---|---|
| `exp(x)` | `eml(x, 1)` |
| `ln(x)` | `eml(1, eml(eml(1, x), 1))` |
| `sin`, `cos`, `π`, `i`, … | deeper compositions |

Every EML expression is a binary tree. The grammar is trivially simple:

```
S → 1 | eml(S, S)
```

This structural regularity makes EML a plausible substrate for symbolic regression, compiler-style normalization, and eventual export to proof-oriented systems.

---

## What This Project Does

`eml-parser` builds a toolchain around EML and uses it as an apparatus for oracle-controlled symbolic regression experiments. The architecture has two cleanly separated layers.

### Raw EML Layer

The parser, AST, and evaluators own the atomic EML language. This layer understands only:

- the constant `1`
- variables
- `eml(left, right)` application

The parser is generated from `parser/eml.y` using `goyacc`. The AST is designed for evaluation, normalization, tree rewriting, serialization, and eventual export to proof tooling.

Two evaluation backends are provided:

- **Fast backend** — `complex128` arithmetic for broad screening, using principal-branch complex `exp` and `log`.
- **High-precision backend** — `big.Float`-backed complex values via `github.com/mshafiee/bigmath` for trusted validation. Currently stable for a subset of the concept library; known edge cases are tracked in `BACKLOG.md`.

### Concept Dictionary Layer

Named mathematical concepts live in a separate registry, not in the raw grammar. Each concept is a parameterized definition that may reference lower-level concepts. Expansion is recursive: every concept ultimately reduces to a raw EML tree composed only of `1`, variables, and `eml(left, right)`.

The current standard library covers:

| Category | Concepts |
|---|---|
| Constants | `one`, `zero`, `e`, `minus_one`, `two`, `half`, `i`, `pi` |
| Exp / log | `exp`, `log`, `id` |
| Arithmetic | `sub`, `neg`, `add`, `recip`, `mul`, `div`, `pow`, `square`, `sqrt` |
| Hyperbolic | `sinh`, `cosh`, `tanh` |
| Trigonometric | `sin`, `cos`, `tan` |

For example, `tan` is not parser syntax — it is a dictionary key whose body references `sin` and `cos`, which in turn reduce further until only raw EML remains. The fully expanded form is a raw EML tree indistinguishable from any other.

---

## Experiment Methodology

The primary goal of the project at this stage is **oracle-controlled symbolic regression experiments**, not open-ended discovery.

Every experiment starts from a known target law and asks: can the current EML tooling recover the intended structure from data, under explicit search bounds and sampling conditions?

### Study Types

- **Positive controls** — targets the current system should recover (e.g. `exp`, `log`).
- **Negative controls** — targets not expected to recover under the current search regime, used to confirm honest failure.
- **Stretch controls** — targets near the edge of current capability, used to map the recovery boundary.

### Recovery Classes

Every experiment result is assigned exactly one class, applied in priority order:

1. `exact_normalized_recovery` — the top candidate matches the expected normalized canonical key exactly.
2. `concept_equivalent_recovery` — the top candidate matches a declared acceptable equivalent key.
3. `approximate_only_recovery` — no structural match, but the top candidate meets a declared numeric score threshold.
4. `no_recovery` — none of the above.

Numeric closeness is never treated as equivalent to structural recovery.

### Current Results

The initial oracle suite shows:

- `exp` and `log` are exactly recoverable under the current bounded enumerative real search.
- A small nested composite (`exp(exp(x))`) is also exactly recoverable.
- `sin`, `sigmoid`, and additive composites beyond the small exact regime are currently honest `no_recovery` failures — the search boundary, not a flaw in EML itself.

---

## Repository Layout

```
ast/                    Raw EML AST types
concepts/               Concept registry, standard library, expansion engine
eval/                   Evaluation backends (complex128 and high-precision)
experiment/             Experiment schema, harness, classification, reporting
experiments/
  specs/                Committed experiment spec JSON files (source of truth)
  datasets/             Generated datasets (reproducible, gitignored by default)
  results/              Per-run result artifacts (reproducible, gitignored by default)
  reports/              Suite-level summaries (JSON + Markdown)
cmd/emltool/            CLI tool for inspecting the concept dictionary
parser/                 goyacc grammar (eml.y)
.docs/                  Architecture and methodology documents
```

---

## Known Limitations

Full details are in `BACKLOG.md`. The main open items:

1. **Inverse function branch semantics** — `asin` and `atanh` return the wrong imaginary sign for some branch-cut inputs. Branch-sensitive validation is currently restricted to `log`, `sqrt`, and `acosh`.
2. **High-precision backend edge cases** — intermediate zero / branch-cut compositions can produce `logarithm undefined for zero` in the `bigmath` backend. High-precision validation is currently restricted to the concepts known to be stable.
3. **Search strategy** — the current search is bounded and enumerative with no constant fitting or gradient refinement. Negative results may reflect search limits rather than EML expressiveness.

---

## Long-Term Direction

The longer-term pipeline this project is designed to support:

1. Infer candidate expressions from data or search.
2. Represent them as normalized raw EML trees.
3. Evaluate with fast and high-precision backends.
4. Export normalized forms to a proof-oriented system such as Lean for verification.

The runtime is currently Go. A future migration toward Zig for the numeric core and search-heavy layers is documented in `.docs/runtime-strategy.md` as a possibility once Go's `big.Float` interop becomes the dominant bottleneck.

---

## What The Project Can Credibly Claim Today

The paper-readiness notes (`.docs/paper-readiness.md`) are explicit about this. The strongest defensible framing right now is:

> A minimal raw EML substrate plus a concept expansion layer plus bounded enumerative search can support reproducible oracle recovery experiments, with results distinguishing exact structural recovery from approximate-only matches and from honest failure — while identifying the current recovery boundary and the major missing pieces needed for broader claims.

This is narrower than a full symbolic regression paper, but it is much more defensible.
