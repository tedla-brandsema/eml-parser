# Initial Oracle Suite

This suite establishes the first empirical baseline for the current
oracle-controlled EML workflow.

Included experiments:

- `oracle_exp_exact.json`
  - exact recovery control for `exp(x)`
- `oracle_log_exact.json`
  - exact recovery control for `log(x)` with a larger node budget
- `oracle_exp_exp_exact.json`
  - exact nested composite control for `exp(exp(x))`
- `oracle_add_exp_x_negative.json`
  - additive composite control expected to fail under current bounded search
- `oracle_sin_negative.json`
  - negative control for `sin(x)`
- `oracle_sigmoid_negative.json`
  - negative control for `sigmoid(x)`

Rationale:

- exact controls show the current search can recover small target laws
- the nested composite shows at least one small composed law is still reachable
- the additive and larger-library controls map the current failure boundary

Current intended reporting flow:

1. run each spec with `emltool run-experiment`
2. aggregate the result files with `emltool report-suite`

Example:

```bash
go run ./cmd/emltool run-experiment experiments/specs/oracle_exp_exact.json
go run ./cmd/emltool run-experiment experiments/specs/oracle_log_exact.json
go run ./cmd/emltool run-experiment experiments/specs/oracle_exp_exp_exact.json
go run ./cmd/emltool run-experiment experiments/specs/oracle_add_exp_x_negative.json
go run ./cmd/emltool run-experiment experiments/specs/oracle_sin_negative.json
go run ./cmd/emltool run-experiment experiments/specs/oracle_sigmoid_negative.json

go run ./cmd/emltool report-suite initial_oracle_suite \
  experiments/results/oracle_exp_exact.json \
  experiments/results/oracle_log_exact.json \
  experiments/results/oracle_exp_exp_exact.json \
  experiments/results/oracle_add_exp_x_negative.json \
  experiments/results/oracle_sin_negative.json \
  experiments/results/oracle_sigmoid_negative.json
```
