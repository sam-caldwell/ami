# AMI Linter (ami lint)

The AMI linter helps maintain code quality and catch common mistakes early. It operates in two layers:

- Stage A: workspace and source scans that do not require full parsing.
- Stage B: parser‑backed rules that inspect `.ami` files for deeper issues.

## Usage

- Human output: `ami lint`
- JSON output: `ami lint --json`
- Strict mode (warnings → errors): `ami lint --strict`
- Verbose debug (writes `build/debug/lint.ndjson`): `ami lint --verbose`

Stage B toggles (advanced rules):
- `--stage-b` enables Stage B.
- `--rule-memsafe` memory safety diagnostics (`E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_BLOCK_UNSUPPORTED`, `E_MUT_ASSIGN_UNMARKED`).
- `--rule-raii` RAII hint (`W_RAII_OWNED_HINT`).
- `--rule-unused` unused imports for identifier‑form imports (`W_UNUSED_IMPORT`).

## Rules (selected)

- Workspace
  - `E_WS_MISSING`, `E_WS_PARSE`, `E_WS_SCHEMA` basic workspace checks.
  - `W_IMPORT_SYNTAX`, `W_IMPORT_ORDER`, `W_IMPORT_DUPLICATE` import shape/order.
  - `W_IMPORT_LOCAL_MISSING`, `W_IMPORT_LOCAL_UNDECLARED`, `E_IMPORT_CONSTRAINT` local path and version checks.
  - `W_PKG_VERSION_INVALID` invalid SemVer in package declarations.
  - `E_IMPORT_CYCLE` circular local-import references among workspace packages (./path).

- Source
  - `W_UNKNOWN_IDENT` sentinel detection for scaffolding.
  - `W_IDENT_UNDERSCORE` underscores in identifiers (use camelCase/PascalCase).
  - `W_LANG_NOT_GO`, `W_GO_SYNTAX_DETECTED` language reminder and Go‑syntax detection.

- Stage B (parser‑backed)
  - Memory safety: `E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_BLOCK_UNSUPPORTED`, `E_MUT_ASSIGN_UNMARKED`.
  - RAII hint: `W_RAII_OWNED_HINT` when `release(x)` is not wrapped in `mutate(...)`.
  - Unused imports: `W_UNUSED_IMPORT` for identifier‑form imports.

## Configuration

In `ami.workspace` you can tune severities and suppressions:

```
toolchain:
  linter:
    rules:
      W_IDENT_UNDERSCORE: info
      W_IMPORT_ORDER: off
    suppress:
      - path: ./vendor
        codes: ["W_*"]
```

Inline per‑file suppression is supported via pragmas:

```
#pragma lint:disable W_IDENT_UNDERSCORE,E_PTR_UNSUPPORTED_SYNTAX
```

## Examples

Circular local import cycle (error)

- Human output (non‑JSON):

```
lint: error E_IMPORT_CYCLE: circular local import detected (ami.workspace)
lint: 1 error(s), 0 warning(s)
```

- JSON (NDJSON) output:

```
{"schema":"diag.v1","timestamp":"2025-01-01T00:00:00Z","level":"error","code":"E_IMPORT_CYCLE","message":"circular local import detected","file":"ami.workspace","data":{"cycle":["./a","./b","./c"]}}
{"schema":"diag.v1","timestamp":"2025-01-01T00:00:00Z","level":"info","code":"SUMMARY","message":"lint summary","data":{"errors":1,"warnings":0}}
```

Notes:
- Cycles are detected only for local imports that refer to declared workspace packages (e.g., `./a`).
- The cycle list is canonicalized (rotated) to start at the lexicographically smallest node and is de‑duplicated.
- Any error (including `E_IMPORT_CYCLE`) causes a non‑zero exit from `ami lint`.
