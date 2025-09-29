# Merge Runtime Planner (Scaffold)

This note defines how normalized `merge.*` attributes on `Collect` map to a runtime merge operator plan. It documents
contracts and determinism requirements for the planner. Implementation is deferred until runtime integration, per
roadmap.

Scope and Goals

- Input: normalized attributes from pipelines debug (`mergeNorm`):
  - `buffer{capacity,policy}`, `stable`, `sort[{field,order}]`, `key`, `partitionBy`, `timeoutMs`, `window`,
    `watermark{field,lateness}`, `dedup`.
- Output: a deterministic operator plan with explicit, validated fields suitable for a runtime executor.
- No behavior is implemented in this phase. This is a design/contract document ensuring stable mapping and no ambiguity.

Operator Plan Schema (draft)

- Name: `merge.plan.v1`
- Fields:
  - `partition`:
    - `by`: string (empty means no partitioning). If both `key` and `partitionBy` are present and differ, this is an
      error (already enforced by semantics).
  - `key`: string (primary key field for deduplication, if any). Missing means dedup uses event identity or is disabled.
  - `ordering`:
    - `fields`: array of `{ field, order }` (order in {asc,desc}); empty means FIFO arrival order.
    - `stable`: bool; true ties break by arrival order.
  - `buffer`:
    - `capacity`: int ≥0
    - `policy`: `block|dropOldest|dropNewest` (legacy alias `drop` rejected at runtime; linter warns)
  - `timeout`:
    - `ms`: int ≥0 (0 means disabled)
  - `window`:
    - `size`: int ≥0 (0 means disabled)
  - `watermark`:
    - `field`: string (required when watermark enabled)
    - `lateness`: string (duration-like literal, normalized by runtime)
  - `dedup`:
    - `field`: string (empty means use `key` if present; otherwise disabled)

Determinism

- Planner must produce byte-for-byte stable JSON for identical inputs.
- Field ordering in arrays (`ordering.fields`) preserves specified order.
- No map iteration in public artifacts; use fixed struct fields and slices.

Validation (input already guarded by semantics)

- Conflicts: `partitionBy` vs `key` mismatch is an error (semantics: `E_MERGE_ATTR_CONFLICT`).
- Requireds: `watermark` requires a field; `sort` requires field; arity checks enforced by parser/semantics.
- Smells: tiny buffer with drop policy, non-positive window/lateness remain lints (non-fatal) unless strict mode
elevates.

Mapping Rules

- `merge.Key(f)`: sets `key=f`.
- `merge.PartitionBy(f)`: sets `partition.by=f` (and must equal `key` if both present).
- `merge.Dedup([f])`: sets `dedup.field = f` (or `key` when empty and `key` non-empty).
- `merge.Sort(f[,ord])`: appends `{field:f, order:ord|asc}` to `ordering.fields`.
- `merge.Stable()`: sets `ordering.stable=true`.
- `merge.Buffer(n[,policy])`: sets `buffer.capacity=n`, `buffer.policy` if present.
- `merge.Timeout(ms)`: sets `timeout.ms=ms`.
- `merge.Window(n)`: sets `window.size=n`.
- `merge.Watermark(f, lateness)`: sets `watermark.field=f`, `watermark.lateness=lateness`.

Interop and Future Work

- Planner hooks: `ami runtime` will ingest `merge.plan.v1` per Collect node.
- Execution semantics (ordering, buffering, watermarks) will be defined in the runtime docs when implemented.
- Cross-pipeline composition and capabilities are outside this scope.

Examples

```
// Collect(merge.Buffer(4, dropOldest), merge.Stable(), merge.Sort(ts, asc), merge.Key("id"))
{
  "schema": "merge.plan.v1",
  "partition": {"by": ""},
  "key": "id",
  "ordering": {"fields": [{"field": "ts", "order": "asc"}], "stable": true},
  "buffer": {"capacity": 4, "policy": "dropOldest"},
  "timeout": {"ms": 0},
  "window": {"size": 0},
  "watermark": null,
  "dedup": {"field": ""}
}
```

Ordering Keys and Typing

- `merge.Sort("field")` uses a payload field as an ordering key. Deep paths like `a.b.c` are supported for `Struct{}` payloads.
- When the upstream payload is a primitive (e.g., `Event<int>`), referencing a field is an error.
- Optional/Union:
  - If the resolved leaf is `Optional<T>`, ordering uses `T` and propagates optionality; missing values are handled by runtime policy.
  - If the resolved leaf is `Union<...>`, ordering is allowed only if all alternatives are orderable primitives.
  - Container and `Struct` leaves are not orderable.

Diagnostics (semantics layer)

- `W_MERGE_FIELD_UNVERIFIED`: No upstream type information to verify the field.
- `E_MERGE_FIELD_ON_PRIMITIVE`: Field used but upstream payload is a primitive type.
- `E_MERGE_SORT_FIELD_UNKNOWN`: Field not found on typed upstream payload(s).
- `E_MERGE_SORT_FIELD_UNORDERABLE`: Field exists but is not orderable (container/struct or mixed union).

Status

- Mapping and contracts documented; implementation deferred to runtime integration milestone per SPEC.
- Runtime behavior

- Buffer/policy enforced per partition. `capacity=0` means effectively unbounded, but backpressure policy is still parsed.
- Timeout clears idle partition buffers older than the timeout when `ExpireStale(now)` is invoked by the scheduler.
- Watermark drops events older than `now - lateness` based on the configured field.
- Window (if set) further bounds in-flight items beyond `buffer.capacity`.
