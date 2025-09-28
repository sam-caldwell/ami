# Logger (logging package)

The `ami/logging` package provides deterministic structured logging with two renderers:

- JSON lines renderer with stable key ordering and ISO-8601 UTC timestamps with milliseconds.
- Human renderer with optional ANSI colors; when `Verbose` is enabled, each line is prefixed with a timestamp.

Key properties:
- CRLF is normalized to LF for deterministic output.
- Multi-line messages in human mode receive a per-line prefix when `Verbose` is true.
- Colors only apply in human mode and when `Color=true`.
- In verbose mode, logs are also appended to `build/debug/activity.log` in addition to the primary writer.

## Usage

```go
import (
    "os"
    "github.com/sam-caldwell/ami/src/ami/logging"
)

func example() error {
    lg, _ := logging.New(logging.Options{
        JSON:     false,      // set true for JSON lines
        Verbose:  true,       // prefix human lines with timestamp; also write build/debug/activity.log
        Color:    true,       // colors only used in human mode
        Package:  "example/app",
        Out:      os.Stdout,  // optional; defaults to stdout
        DebugDir: "build/debug", // optional; default if empty
    })
    defer lg.Close()

    lg.Info("hello world", map[string]any{"count": 1})
    lg.Warn("multi\nline", nil)
    lg.Error("boom", map[string]any{"err": "E42"})
    return nil
}
```

## JSON Schema

The JSON renderer emits a stable object per line with keys in the order:
`timestamp`, `level`, `package` (optional), `message`, `fields` (optional, keys sorted).

Example line:

```json
{"timestamp":"2025-09-24T17:05:06.123Z","level":"info","package":"example/app","message":"hello","fields":{"a":1}}
```

## Human Renderer

Example (with `Verbose=true`, `Color=true`):

```text
2025-09-24T17:05:06.123Z [INFO] example/app: hello
2025-09-24T17:05:06.124Z [WARN] example/app: first line
2025-09-24T17:05:06.124Z [WARN] example/app: second line
```

## Pipeline

Verbose CLI logs flow through the stdlib logger Pipeline and are written to `build/debug/activity.log` via a file sink. Primary command outputs (stdout/stderr) remain unchanged.

Defaults chosen to preserve current behavior:
- `capacity`: 256
- `batchMax`: 1 (line-by-line)
- `flushInterval`: 0 (disabled)
- `backpressure`: `block`

Backpressure policies supported:
- `block`: producer blocks until there is capacity
- `dropNewest`: drop the new item when full
- `dropOldest`: drop one oldest queued item to make room

Flushing:
- Batches flush when `batchMax` items are buffered or when `flushInterval` elapses.
- On `Close()`, the pipeline drains the queue and flushes a final batch.

Redaction safety-net (JSON):
- Pipeline applies an additional redaction pass when configured, even for preformatted `log.v1` lines.
- Order: allowlist → denylist → redact exact keys → redact by prefix.

Counters (observability):
- `enqueued`, `written`, `dropped`, `batches`, `flushes` (via `Stats()`)
- Deterministic and test-friendly; no random IDs; timestamps remain only inside log records.

## Notes
- Avoid writing logs to stdout when CLI commands emit JSON payloads (tests expect clean JSON). Prefer writing to `build/debug/activity.log` in verbose mode for debugging.
- M3 milestone wires CLI debug logs to the stdlib pipeline; remote or async sinks are out of scope for M3.

### Redaction and Field Filters

When running CLI commands with `--verbose`, you can mask or filter fields in debug logs:

Examples:

```
# Redact exact keys
ami --verbose --redact target --redact password init --json --force

# Redact by prefix (e.g., any key starting with "meta.")
ami --verbose --redact-prefix meta. lint --json

# Allowlist only specific keys; deny specific ones
ami --verbose --allow-field target --deny-field force --deny-field json init --json --force
```

- `--redact` masks exact field keys with `[REDACTED]`.
- `--redact-prefix` masks any field whose key starts with one of the given prefixes.
- `--allow-field` keeps only the listed keys (if provided); others are dropped.
- `--deny-field` drops listed keys after allow filtering.
- Redaction and filtering happen in the logger before formatting, so JSON logs remain deterministic.
- These affect debug logs only (e.g., `build/debug/activity.log`), not primary command outputs.

Note: For eventual pipeline adoption, the stdlib logger pipeline also applies a safety-net redaction pass for `log.v1` JSON lines when configured.
