# Observability (1.6): Event IDs and Hooks

Event Lifecycle Metadata

- The compiler emits `eventmeta.v1` debug artifacts with standardized fields:
  - `id` (string), `timestamp` (ISO‑8601), `attempt` (int)
  - `trace.traceparent`, `trace.tracestate` following W3C Trace Context.
- See `build/debug/ir/<package>/<unit>.eventmeta.json` during `ami build --verbose`.

Telemetry Hooks

- Telemetry configuration is not expressed via language pragmas. Hooks and integration points are configured via tooling/runtime, not source tokens.

Notes

- This is a scaffold for hooks presence and configuration. Integration with a runtime/SDK can bind these tokens to specific instrumentation backends.
