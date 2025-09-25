# Observability (Ch. 1.6)

Implemented observability features provide:

- Telemetry pragma: captured in ASM header (e.g., `; telemetry trace,counters`).
- Event metadata schema (`eventmeta.v1`): trace context fields (`traceparent`, `tracestate`).
- Metrics hooks: runtime helpers to emit pipeline/node metrics as `diag.v1` JSON lines.

Usage

- Enable JSON logs (CLI sets this for `--json`), or in code via `logger.Setup(true, ...)`.
- Emit metrics from runtime code:
  - `metrics.PipelineMetrics{Pipeline:"P", QueueDepth:7, Throughput:123.4, LatencyMs:20, Errors:1}.Emit()`
  - `metrics.NodeMetrics{Pipeline:"P", Node:"Transform", QueueDepth:3, Throughput:42.0, LatencyMs:5, Errors:0}.Emit()`
- Output format (diag.v1):
  - `{ "schema":"diag.v1", "timestamp":"...", "level":"info", "message":"metrics.pipeline", "data":{...} }`

Notes

- These are scaffolds for consistent, machine‑readable logs; they do not implement a metrics backend.
- Field names are stable and intended for tooling ingestion.
- Event metadata and telemetry pragma are emitted in debug artifacts during `ami build --verbose`.
