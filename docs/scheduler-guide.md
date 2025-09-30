# Scheduler Guide

How workers, limits, and policies interact; with backpressure examples.

## Workers and Schedule
- `#pragma concurrency:workers N` sets the default worker pool size (N>=1).
- `#pragma concurrency:schedule <fifo|lifo|fair|worksteal>` picks a scheduling policy.
- Lints:
  - `E_CONCURRENCY_WORKERS_INVALID`: workers < 1.
  - `E_CONCURRENCY_SCHEDULE_INVALID`: unknown policy.
  - `W_CONCURRENCY_SCHEDULE_IGNORED`: policy likely ineffective when `workers=1`.
  - `W_CONCURRENCY_SCHEDULE_UNSPECIFIED`: workers>1 but schedule omitted.

## Per-Kind Limits
- `#pragma concurrency:limits ingress=... transform=... fanout=... collect=... mutable=... egress=...`
- Each value must be >=1; unknown keys are rejected.
- Lints:
  - `W_CONCURRENCY_LIMITS_UNSPECIFIED`: no limits declared.
  - `W_CONCURRENCY_LIMIT_UNUSED`: limit declared for a kind not present in pipelines.

## Backpressure and Buffers
- Step attributes, typically on `Collect`, control buffering and delivery:
  - `merge.Buffer(n[,policy])`: capacity (n>=0); policies: `block|dropOldest|dropNewest`.
  - `merge.Timeout(ms)`: positive integer milliseconds.
  - `merge.Window(n)`: non-negative size (0 disables).
  - `merge.Watermark(field, lateness)`: lateness as positive int or duration (e.g., `100ms`, `1s`).
- Lints and validations:
  - `E_MERGE_ATTR_TYPE`: wrong types (e.g., non-integer for Timeout/Window).
  - `E_MERGE_ATTR_ARGS`: invalid values (e.g., Timeout must be >0).
  - `W_MERGE_TINY_BUFFER`: tiny buffer with dropping policy.
  - `W_MERGE_WINDOW_ZERO_OR_NEGATIVE`, `W_MERGE_WATERMARK_NONPOSITIVE`.

## Delivery Semantics
- Delivery is derived from buffer policy for contracts/debug:
  - `block` → atLeastOnce
  - `dropOldest`/`dropNewest` → bestEffort

## Example
```
#pragma concurrency:workers 4
#pragma concurrency:schedule worksteal
#pragma concurrency:limits transform=8 collect=4

pipeline P(){
  ingress; Transform(A);
  Collect merge.Buffer(64, dropOldest), merge.Stable(), merge.Sort(ts, asc);
  egress;
  ingress -> Transform; Transform -> Collect; Collect -> egress;
}
```

