# Merge (Collect + edge.MultiPath)

This document seeds the merge/collect design for AMI, summarizing semantics and examples that align with the Asynchronous Machine Interface guidance. It describes how to express multi‑upstream merge behavior on `Collect` nodes using `edge.MultiPath(...)` and `merge.*(...)` attributes.

Status: design seeded for implementation. The `edge.MultiPath` facility and `merge` attributes are specified here and listed in SPECIFICATION.md; code integration proceeds in phases.

## Overview

- `Collect` may merge multiple upstream paths into a single downstream stream.
- `edge.MultiPath(...)` is declared on the `Collect` step to configure merge behavior.
- Attributes are expressed as `merge.*(...)` calls inside the `edge.MultiPath(...)` list. Mixed `k=v` pairs and `merge.*(...)` are permitted; attribute order is not significant; last‑write wins on duplicates.

## Supported Attributes (planned)

- `merge.Sort(field[, order])`: define output ordering by payload field. `order` ∈ {`asc`, `desc`} (default `asc`). See “Sort Semantics”.
- `merge.Stable()`: request stability for equal keys; preserves arrival order inside the active merge window/partition.
- `merge.Key(field)`: define a key field used by other attributes (e.g., `merge.Dedup`).
- `merge.Dedup([field])`: drop duplicates based on the keyed field (or `merge.Key` if omitted).
- `merge.Window(size)`: bound in‑flight merge window.
- `merge.Watermark(field, lateness)`: watermark scheduling; tolerate out‑of‑order arrival by `lateness`.
- `merge.Timeout(ms)`: upper bound on how long a window remains open before emission.
- `merge.Buffer(capacity[, backpressure])`: internal buffer and backpressure policy (`block` or `drop`).
- `merge.PartitionBy(field)`: partition by key; merges/sorts independently per partition.

## Sort Semantics (Collect)

- Field selection
  - `field` is a selector into the payload (e.g., `event.ts`, `payload.id`, `meta.trace.id`). Dotted selectors allowed.
  - Field must be present or resolvable for ordering; missing/null values sort after present values (nulls‑last) unless overridden in future.
- Order argument
  - Optional: `asc` (default) or `desc`.
- Type‑specific ordering (deterministic)
  - Integer/float: numeric; NaN (if representable) sorts after numbers.
  - Boolean: `false < true` for asc (inverted for desc).
  - Timestamp: normalized (e.g., epoch ns) compare; independent of timezone/locale.
  - String: binary UTF‑8 comparison; no collation.
- Windowing & watermarks
  - Sorting applies within the active merge window only, as determined by `merge.Window`, `merge.Timeout`, and/or `merge.Watermark`.
  - Watermarks advance time and flush windows; late arrivals are handled per the configured policy (drop/next window), deterministically.
- Partitioning
  - When `merge.PartitionBy(field)` is present, sorting applies independently per partition key.
- Stability & tiebreaks
  - With `merge.Stable()`: stable sort; equal keys preserve arrival order within the window/partition.
  - Without `merge.Stable()`: sort may be unstable but MUST remain deterministic for identical inputs.
  - Secondary tiebreak MAY use `merge.Key(field)` (if orderable); otherwise (partition, upstreamIndex, arrivalIndex) acts as a deterministic fallback.
- Buffer/backpressure interaction
  - `merge.Buffer(capacity, backpressure)` constrains sorting memory.
  - `backpressure=block`: upstreams block; ordering remains per window.
  - `backpressure=drop`: records may be dropped on overflow; output remains sorted among retained records.

## Examples

Note: examples are illustrative and show intended AMI syntax. Normal/verbose whitespace and quoting are for clarity.

### 1) Time‑sorted Collect with stable ordering and watermark

```
package main

func parse(ctx Context, ev Event<string>, st State) Event<Log> {}
func sink(ctx Context, ev Event<Log>, st State) Ack {}

pipeline Logs {
  Ingress(cfg)
    .Transform(parse)
    .FanOut(http, syslog)
    .Collect(
      in=edge.MultiPath(
        merge.Sort("event.ts", "asc"),
        merge.Stable(),
        merge.Watermark("event.ts", "5s"),
        merge.Buffer(1024, backpressure="block")
      )
    )
    .Egress(sink)
}
```

- Sort by `event.ts` ascending, stable ordering for ties.
- Watermark tolerates 5s out‑of‑order; windows close when watermark advances.
- Buffer 1024 events, block on capacity.

### 2) Descending severity with bounded drop buffer

```
package main

func normalize(ctx Context, ev Event<string>, st State) Event<Alert> {}
func alertSink(ctx Context, ev Event<Alert>, st State) Ack {}

pipeline Alerts {
  Ingress(cfg)
    .Transform(normalize)
    .Collect(
      in=edge.MultiPath(
        merge.Sort("payload.severity", "desc"),
        merge.Buffer(256, backpressure="drop")
      )
    )
    .Egress(alertSink)
}
```

- Higher severity first. If the buffer is full, least recent arrivals may be dropped; retained records are still properly ordered.

### 3) Partitioned merge by tenant with per‑partition sort by id

```
package main

func decode(ctx Context, ev Event<string>, st State) Event<Record> {}
func out(ctx Context, ev Event<Record>, st State) Ack {}

pipeline TenantMerge {
  Ingress(cfg)
    .Transform(decode)
    .Collect(
      in=edge.MultiPath(
        merge.PartitionBy("meta.tenant"),
        merge.Sort("payload.id", "asc"),
        merge.Stable(),
        merge.Window(1000),
        merge.Timeout(50)
      )
    )
    .Egress(out)
}
```

- Partitions by tenant, sorts by `payload.id` ascending within each partition.
- Stable sort with a window bound and a timeout for regular emission.

## Notes

- Attribute precedence: later attributes in the list override earlier ones for the same setting.
- Determinism: identical inputs and configuration MUST yield identical outputs across runs.
- Planner/runtime mapping: see SPECIFICATION.md for IR fields; implementation will map attributes to merge operator configuration in later phases.

