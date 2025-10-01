# Runtime Test Harness

Overview of `ami test` runtime execution support, KV store integration, and error pipeline emission.

- CLI flags: `--kv-metrics`, `--kv-dump`, `--kv-events`, per-case `emit=true`.
- Artifacts: process-level under `build/test/kv/` and per-case under `build/test/kv/<file>_<case>.*.json`.
- Default ErrorPipeline on error cases emits `errors.v1` JSON lines to stderr; toggles: `--no-errorpipe`, `--errorpipe-human`.

See also:
- `docs/toolchain/runtime-kvstore.md` — KV store design and API
- `docs/toolchain/pipelines-v1-quickstart.md` — debug artifact navigation
