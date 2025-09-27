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
- `--rule-memsafe` — memory safety diagnostics (`E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_BLOCK_UNSUPPORTED`, `E_MUT_ASSIGN_UNMARKED`).
- `--rule-raii` — hint to wrap `release(x)` calls (`W_RAII_OWNED_HINT`).
- `--rule-unused` — report unused identifier‑style imports (`W_UNUSED_IMPORT`).

What gets checked (high level):
- Workspace: `ami.workspace` exists and parses; has a version and a `main` package. Local imports are valid; duplicate or cyclic imports are flagged.
- Source style: underscores in identifiers, language reminders (`W_LANG_NOT_GO`, `W_GO_SYNTAX_DETECTED`), TODO/FIXME policy.
- Stage B: memory‑safety, RAII hints, unused imports, simple pipeline checks (ingress/egress placement, unreachable nodes), and basic buffer/backpressure smells.

Exit codes:
- Success: 0.
- Failure: non‑zero when there are errors, or when `--strict` promotes warnings, or when `--failfast`/`--max-warn` triggers.

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
