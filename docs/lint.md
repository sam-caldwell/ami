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
- Merge/multipath: `E_MERGE_ATTR_REQUIRED`, `E_MERGE_ATTR_ARGS`, `E_MERGE_ATTR_CONFLICT`, `W_MERGE_SORT_NO_FIELD`, `W_MERGE_FIELD_UNVERIFIED`, `E_MERGE_SORT_FIELD_UNKNOWN`, `E_MERGE_SORT_FIELD_UNORDERABLE`.

Examples

```
// memory safety
a := &b                  // E_PTR_UNSUPPORTED_SYNTAX
*c + d                   // E_MUT_BLOCK_UNSUPPORTED
func F(){ x = 1 }        // E_MUT_ASSIGN_UNMARKED

// RAII hint
func G(){ release(x) }   // W_RAII_OWNED_HINT; prefer mutate(release(x))

// pipeline smells
pipeline P(){ Collect.buffer(1, drop); egress; }  // W_PIPELINE_BUFFER_DROP_ALIAS
pipeline Q(){ ingress(); work(); egress(); tail } // W_PIPELINE_EGRESS_POS
```

What gets checked (high level):
- Workspace: `ami.workspace` exists and parses; has a version and a `main` package. Local imports are valid; duplicate
or cyclic imports are flagged.
- Source style: underscores in identifiers, language reminders (`W_LANG_NOT_GO`, `W_GO_SYNTAX_DETECTED`), TODO/FIXME
policy.
- Stage B: memory‑safety, RAII hints, unused imports, simple pipeline checks (ingress/egress placement, unreachable
nodes), and basic buffer/backpressure smells.

Exit codes:
- Success: 0.
- Failure: non‑zero when there are errors, or when `--strict` promotes warnings, or when `--failfast`/`--max-warn`
triggers.

Filtering by rule code (`--rules`):
- Substring: `IDENT` matches `W_IDENT_UNDERSCORE`.
- Glob: `W_PIPELINE_*` matches all pipeline rules.
- Regex: `re:^W_(IDENT|PIPELINE)_` or `/^E_IMPORT_/`.

Configuring severity and suppressions (`ami.workspace`):

```
toolchain:
  linter:
    rules:
      W_IDENT_UNDERSCORE: info   # lower severity
      W_IMPORT_ORDER: off        # disable rule
    suppress:
      - path: ./vendor           # ignore vendor/
        codes: ["W_*"]
```

Inline per‑file suppression (top of the file):

```
#pragma lint:disable W_IDENT_UNDERSCORE,E_PTR_UNSUPPORTED_SYNTAX
```

Example: circular local import (error)

- Human output:

```
lint: error E_IMPORT_CYCLE: circular local import detected (ami.workspace)
lint: 1 error(s), 0 warning(s)
```

- JSON (NDJSON):

```
{"schema":"diag.v1","timestamp":"2025-01-01T00:00:00Z","level":"error","code":"E_IMPORT_CYCLE","message":"circular local import detected","file":"ami.workspace","data":{"cycle":["./a","./b","./c"]}}
{"schema":"diag.v1","timestamp":"2025-01-01T00:00:00Z","level":"info","code":"SUMMARY","message":"lint summary","data":{"errors":1,"warnings":0}}
```

Troubleshooting:
- "workspace not found": run `ami init` in your project root to create `ami.workspace`.
- No output: when there are no findings, human mode prints `lint: OK`; JSON mode prints only the final `SUMMARY` record.
- Want fewer findings: adjust `toolchain.linter.rules` or add `#pragma lint:disable` pragmas as shown above.
