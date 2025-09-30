# Merge Field Diagnostics

Guidance and examples for merge.Sort/Key/PartitionBy/Dedup field validation.

## Struct and Nested Fields
- Valid names: `[A-Za-z_][A-Za-z0-9_\.]*` (dot-separated parts).
- Example:
```
// Sort by a nested timestamp
Collect merge.Sort("meta.ts", asc)
```

## Missing Field
- Symptom: `E_MERGE_ATTR_REQUIRED` or `W_MERGE_SORT_NO_FIELD` for `merge.Sort()` without a field.
- Fix: specify a non-empty field name.

## Invalid Field Name
- Symptom: `E_MERGE_FIELD_NAME_INVALID`.
- Cause: invalid identifier or empty part in a dotted path.
- Example (error):
```
Collect merge.Sort(".ts")
```

## Unorderable Field (typed flow)
- Symptom: `E_MERGE_SORT_FIELD_UNORDERABLE` (when type info is available).
- Cause: field resolves to a non-orderable type (container/struct/union with mixed non-orderables).
- Fix: pick a primitive/ordered field or pre-transform data.

## Field On Primitive Payload
- Symptom: `E_MERGE_FIELD_ON_PRIMITIVE`.
- Cause: upstream payload is a primitive (e.g., `Event<int>`) and field access is invalid.
- Fix: remove the field or map to a struct payload first.

## Stability Hints
- `W_MERGE_STABLE_WITHOUT_SORT`: Stable has no effect without Sort.
- `W_MERGE_SORT_POSSIBLY_UNSTABLE`: Sort without Stable and without Key/PartitionBy may be unstable across batches.
- `W_MERGE_STABLE_REDUNDANT`: Stable may be redundant when a unique Key exists alongside Sort.

## Cross-Attribute Alignment
- Sort vs Key:
  - `E_MERGE_SORT_PRIMARY_NEQ_KEY`: when both `merge.Key(k)` and `merge.Sort(f, ...)` are present, the primary sort field must equal `k`.
    - Data: `{ "key": "...", "primary": "...", "sort": ["..."] }`.
- Sort vs PartitionBy (without Key):
  - `E_MERGE_SORT_PRIMARY_NEQ_PARTITION`: when `merge.PartitionBy(p)` is present without `merge.Key`, the primary sort field must equal `p`.
    - Data: `{ "partition": "...", "primary": "...", "sort": ["..."] }`.
- PartitionBy vs Key:
  - `E_MERGE_ATTR_CONFLICT`: conflicting fields between `merge.PartitionBy` and `merge.Key`.
    - Data: `{ "key": "...", "partition": "..." }`.
- Window/Watermark alignment:
  - `W_MERGE_WINDOW_WITHOUT_WATERMARK`: Window present without Watermark; behavior may default to processing time.
    - Data: `{ "window": true }`.
  - `W_MERGE_WATERMARK_NOT_PRIMARY_SORT`: Watermark field differs from the primary Sort field.
    - Data: `{ "watermark": "...", "primary": "...", "sort": ["..."] }`.

Strictness toggle
- Set `AMI_STRICT_DEDUP_PARTITION=1` to elevate `W_MERGE_DEDUP_FIELD_WITHOUT_KEY_UNDER_PARTITION` and
  `W_MERGE_DEDUP_WITHOUT_KEY_UNDER_PARTITION` to their error forms (`E_*`).

## Structured Data (Data field)
Most merge diagnostics include a `data` object for machine-readability. Examples:
- `E_MERGE_ATTR_ARGS`: `{ "argc": 2, "expected_min": 1, "expected_max": 1 }`.
- `E_MERGE_ATTR_TYPE`: attribute-specific key (e.g., `ms`, `size`, `lateness`, `capacity`).
- `E_MERGE_FIELD_NAME_INVALID`: `{ "field": "..." }`.
- `E_MERGE_SORT_ORDER_INVALID`: `{ "order": "..." }`.
- `W_MERGE_TINY_BUFFER`: `{ "capacity": "1", "policy": "dropNewest" }`.
- `W_MERGE_DEDUP_WITHOUT_KEY`: `{ "dedup": true }`.

## Dedup and Partition
- `W_MERGE_DEDUP_WITHOUT_KEY`: Dedup() without a field relies on merge.Key.
- `W_MERGE_DEDUP_WITHOUT_KEY_UNDER_PARTITION`: Dedup() without Key under PartitionBy may be ineffective.
