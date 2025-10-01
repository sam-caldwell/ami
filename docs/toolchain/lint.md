# AMI Linter (ami lint)

The AMI linter checks your project for common mistakes and style issues. It is safe to run early and often.

Two kinds of checks:
- Stage A: fast workspace and source scans (no full parse needed).
- Stage B: deeper, parser‑backed rules for `.ami` files (opt‑in via flags).

Quick start (run from your workspace root):
- `ami lint` — prints human‑readable findings and a summary.
- `ami lint --json` — prints one JSON object per line (NDJSON).
- `ami lint --strict` — treat warnings as errors.

Common options:
- `--verbose` — also write `build/debug/lint.ndjson` with all records.
- `--failfast` — exit non‑zero if any warning or error is found.
- `--rules <pattern>` — keep only matching codes; repeat to add more.
- `--max-warn N` — fail when warnings exceed N (non‑strict mode).
- `--compat-codes` — in `--json`, add `data.compatCode = LINT_<CODE>`.

Advanced (Stage B) options:
- `--stage-b` — enable parser‑backed rules.
- `--rule-memsafe` — memory safety diagnostics (`E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_BLOCK_UNSUPPORTED`,
`E_MUT_ASSIGN_UNMARKED`).
- `--rule-raii` — hint to wrap `release(x)` calls (`W_RAII_OWNED_HINT`).
- `--rule-unused` — report unused identifier‑style imports (`W_UNUSED_IMPORT`).
 - `--strict-merge-dedup-partition` — override strictness for `merge.Dedup` under `merge.PartitionBy` without `merge.Key`.
   - When set, elevates related warnings to errors (see `docs/toolchain/merge-field-diagnostics.md`).
   - Alternative configuration: `toolchain.linter.strict_merge_dedup_partition: true` in `ami.workspace`, or env `AMI_STRICT_DEDUP_PARTITION=1`.

Stage B rules (parser‑backed)

- Memory safety:
  - `E_PTR_UNSUPPORTED_SYNTAX`: `&x` is not allowed.
  - `E_MUT_BLOCK_UNSUPPORTED`: unary `*` is not a dereference; only `* name = expr` is allowed.
  - `E_MUT_ASSIGN_UNMARKED`: plain assignment in function bodies must use `* name = expr`.
- RAII hint:
  - `W_RAII_OWNED_HINT`: wrap `release(x)` in `mutate(release(x))` to signal ownership transfer.
- Imports:
  - `W_DUP_IMPORT_ALIAS`: duplicate alias in import declarations.
  - `W_UNUSED_IMPORT`: alias/path imported but not referenced.
- Functions:
  - `W_DUP_FUNC_DECL`: same function declared in multiple files.
- Pipelines:
  - `W_PIPELINE_INGRESS_POS`: `ingress` should be first.
  - `W_PIPELINE_EGRESS_POS`: `egress` should be last.
  - `W_PIPELINE_UNREACHABLE_NODE`: node appears unreachable from `ingress`.
  - `W_PIPELINE_BUFFER_DROP_ALIAS`: ambiguous `drop` backpressure alias; prefer `dropOldest|dropNewest|block`.
  - `W_PIPELINE_BUFFER_POLICY_SMELL`: capacity <= 1 with drop policy is likely ineffective.
  - Semantics parity (selected): `E_PIPELINE_START_INGRESS`, `E_PIPELINE_END_EGRESS`, `E_DUP_EGRESS`, `E_UNKNOWN_NODE`,
    `E_IO_PERMISSION`.

## Common Diagnostics Cheat Sheet

- Duplicate pipeline name: `E_DUP_PIPELINE`.
- Ingress/Egress placement: `E_PIPELINE_START_INGRESS`, `E_PIPELINE_END_EGRESS`, `E_INGRESS_POSITION`, `E_EGRESS_POSITION`, `E_DUP_INGRESS`, `E_DUP_EGRESS`.
- Unknown or forbidden nodes: `E_UNKNOWN_NODE`, `E_IO_PERMISSION`.
- Edge endpoints/direction: `E_EDGE_UNDECLARED_FROM`, `E_EDGE_UNDECLARED_TO`, `E_EDGE_TO_INGRESS`, `E_EDGE_FROM_EGRESS`, `E_PIPELINE_SELF_EDGE`, `W_PIPELINE_DUP_EDGE`.
- Connectivity: `E_PIPELINE_NODE_DISCONNECTED`, `E_PIPELINE_NO_PATH_INGRESS_EGRESS`, `E_PIPELINE_UNREACHABLE_FROM_INGRESS`, `E_PIPELINE_CANNOT_REACH_EGRESS`, `E_PIPELINE_CYCLE`.
- Edge capacity/backpressure: `E_EDGE_CAPACITY_INVALID`, `E_EDGE_CAPACITY_ORDER`, `E_EDGE_BACKPRESSURE`, `W_EDGE_BP_LEGACY_DROP`.
- Merge/multipath: `E_MERGE_ATTR_REQUIRED`, `E_MERGE_ATTR_ARGS`, `E_MERGE_ATTR_TYPE`, `E_MERGE_ATTR_CONFLICT`, `W_MERGE_SORT_NO_FIELD`, `W_MERGE_FIELD_UNVERIFIED`, `E_MERGE_SORT_FIELD_UNKNOWN`, `E_MERGE_SORT_FIELD_UNORDERABLE`.
 - Merge/multipath: `E_MERGE_ATTR_REQUIRED`, `E_MERGE_ATTR_ARGS`, `E_MERGE_ATTR_TYPE`, `E_MERGE_ATTR_CONFLICT`, `W_MERGE_SORT_NO_FIELD`, `W_MERGE_FIELD_UNVERIFIED`, `E_MERGE_SORT_FIELD_UNKNOWN`, `E_MERGE_SORT_FIELD_UNORDERABLE`, `E_MERGE_SORT_PRIMARY_NEQ_KEY`, `E_MERGE_SORT_PRIMARY_NEQ_PARTITION`, `W_MERGE_WINDOW_WITHOUT_WATERMARK`, `W_MERGE_WATERMARK_NOT_PRIMARY_SORT`.
- Edge coverage: `E_EDGES_WITHOUT_INGRESS`, `E_EDGES_WITHOUT_EGRESS` (edges present but missing ingress/egress).
