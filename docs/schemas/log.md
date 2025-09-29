# log.v1 — Structured Log Record

Fields

- `schema`: fixed `log.v1`.
- `timestamp`: ISO‑8601 UTC with millisecond precision.
- `level`: `trace|debug|info|warn|error` (subset used by CLI).
- `message`: short event name (e.g., `build.start`).
- `fields` (optional): object with context fields; deterministic where applicable.

Example

```
{"schema":"log.v1","timestamp":"2025-01-01T00:00:00.000Z","level":"info","message":"build.start","fields":{"json":false}}
```
