# diag.v1 — Diagnostic Record Schema

Fields

- `schema`: fixed `diag.v1`.
- `timestamp`: ISO‑8601 UTC with millisecond precision.
- `level`: `info|warn|error`.
- `code`: stable string code (e.g., `E_WS_SCHEMA`, `W_IMPORT_ORDER`).
- `message`: human summary.
- `file` (optional): path to related file.
- `pos` (optional): `{ line, column, offset }` with 1‑based line/column.
- `data` (optional): structured extras; key order not guaranteed.

Example

```
{"schema":"diag.v1","timestamp":"2025-01-01T00:00:00.000Z","level":"error","code":"E_PARSE","message":"unexpected token","file":"src/main.ami","pos":{"line":3,"column":14,"offset":52}}
```
