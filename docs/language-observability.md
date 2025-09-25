# Observability (1.6): Event IDs and Telemetry Hooks

Event Lifecycle Metadata

- The compiler emits `eventmeta.v1` debug artifacts with standardized fields:
  - `id` (string), `timestamp` (ISO‑8601), `attempt` (int)
  - `trace.traceparent`, `trace.tracestate` following W3C Trace Context.
- See `build/debug/ir/<package>/<unit>.eventmeta.json` during `ami build --verbose`.

Telemetry Hooks

- Enable via pragma: `#pragma telemetry <tokens>` (comma or space separated).
  - Examples: `#pragma telemetry trace,counters` or `#pragma telemetry trace`.
- The IR module records the list, and the assembly header includes a line like:
  - `; telemetry trace,counters`

Notes

- This is a scaffold for hooks presence and configuration. Integration with a runtime/SDK can bind these tokens to specific instrumentation backends.

