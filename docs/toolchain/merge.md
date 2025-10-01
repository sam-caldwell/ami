## Merge Attributes and MultiPath (Collect)

This document summarizes the current scaffold for `edge.MultiPath` on `Collect` nodes and the `merge.*` attributes that
influence merging behavior.

Status: scaffold with linter hints; normalization and runtime mapping are deferred to later phases.

- Context: `edge.MultiPath(...)` is only valid on `Collect` nodes.
- Attributes (recognized names; arity/type checks are minimal):
  - `merge.Sort(field[, order])` where `order ∈ {asc, desc}`; default `asc`.
  - `merge.Stable()` requests a stable ordering when sort keys tie.
  - `merge.Key(field)` defines a key for other operations.
  - `merge.Dedup([field])` removes duplicates; defaults to `merge.Key` when field is omitted.
  - `merge.Window(size)` sets bounded in‑flight window size.
  - `merge.Watermark(field, lateness)` sets lateness tolerance.
  - `merge.Timeout(ms)` sets a max wait for merge decisions.
  - `merge.Buffer(capacity[, backpressure])` sets internal buffer and backpressure policy. Supported policies: `block`,
    `dropOldest`, `dropNewest`. The legacy alias `drop` is allowed but discouraged (linter warns
    `W_PIPELINE_BUFFER_DROP_ALIAS`).

Linter hints (current):

- `W_PIPELINE_BUFFER_POLICY_SMELL`: policy + capacity combination likely to drop excessively (e.g., capacity <= 1 with
dropping policy).
- `W_PIPELINE_BUFFER_DROP_ALIAS`: warns on ambiguous `drop` alias.

Diagnostics:

- Unknown/invalid merge attributes emit `E_MERGE_ATTR_*` codes with positions.
- Hints use `W_*` codes and can be suppressed via config (`toolchain.linter.suppress`) or pragmas (`#pragma lint:disable
CODE`).

Example:

```
pipeline P(){
  A();
  A Buffer(1, dropNewest); // smell (tiny buffer + dropping policy)
  B();
  B Buffer(1, drop);       // alias warns
}
```

See also: `docs/language/edges.md` for the edges index schema and debug artifacts.

## Merge Rules (At a Glance)

- Sort and Key
  - Primary `merge.Sort(field, [order])` must equal `merge.Key(field)` when both are set.
  - Missing sort order is treated as `asc` for conflict checks.

- Partitioning
  - With `merge.PartitionBy(p)` (and no `merge.Key`), the primary `merge.Sort` field must be `p`.
  - `merge.Dedup(field)` under `merge.PartitionBy` without `merge.Key` is discouraged; enable strict mode to treat as error.

- Windowing and Watermarks
  - `merge.Window(n)` without `merge.Watermark(ts, lateness)` may default to processing‑time semantics.
  - Prefer making the watermark field the primary sort key.

- Stability
  - `merge.Stable` has no effect without `merge.Sort`.
  - Sorting without `Key`/`PartitionBy` and without `Stable` may be unstable across batches.

Strictness toggles
- Set `AMI_STRICT_DEDUP_PARTITION=1` to elevate `merge.Dedup(field)` under `PartitionBy` without `Key` from warn to error.

Diagnostics include structured `data` payloads for machine consumers. See docs/diag-codes.md.
