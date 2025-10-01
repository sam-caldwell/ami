# Logger Pipeline (Toolchain Integration)

This document describes how the CLI wires the stdlib logger pipeline for debug logs and redaction.

- Verbose CLI logs flow through the stdlib logger Pipeline and are written to `build/debug/activity.log` via a file sink. Primary command outputs (stdout/stderr) remain unchanged.
- Redaction: when configured, the pipeline applies an allow/deny pass and key/prefix redaction for `log.v1` JSON lines.
- M3 milestone wires CLI debug logs to the stdlib pipeline; remote or async sinks are out of scope for M3.

See also:
- `docs/language/stdlib/logger.md` — stdlib pipeline and sink APIs
