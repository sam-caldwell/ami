# Runtime Test Harness (Phase 2)

This document describes the directive-driven runtime test harness integrated
with `ami test`. It augments Go test results with native AMI cases and a simple
runtime harness for deterministic execution and observability.

## Directives (in `*_test.ami`)

- `#pragma test:case <name>`: defines a runtime test case name.
- `#pragma test:runtime <kv>`: per-file defaults for runtime expectations.
  - `input={...}`: JSON input payload (object) passed to the runtime harness.
  - `output={...}`: expected JSON output payload; deep-equal on normalized JSON.
  - `expect_error=<CODE>`: expect harness to return an error (by code string).
  - `timeout=<ms>`: per-case timeout override in milliseconds.
- `#pragma test:kv <kv>`: key-value configuration for node-state store ops:
  - `ns="<namespace>"`: store namespace (e.g., `pipeline/node`); defaults to the global store.
  - `put="k1=v1;k2=v2"`: semicolon-separated put list applied before execution.
  - `get="k1,k2"`: comma-separated get list applied after execution (side-effects only).
  - `emit=true`: emit per-case KV metrics and dump files.

Multiple cases are allowed; all `test:kv` directives in the file are applied to
all declared cases in the file.

## Artifacts

When runtime cases are present, `ami test` augments outputs:
- JSON mode: streams `test.v1` events (`run_start`, `test_start`, `test_end`, `run_end`).
- Verbose mode: writes `build/test/runtime.log` and appends `runtime:<file> <case>` to `build/test/test.manifest`.
- KV per-case (when `emit=true` or `--verbose`):
  - `build/test/kv/<file>_<case>.metrics.json` (`kv.metrics.v1`)
  - `build/test/kv/<file>_<case>.dump.json` (`kv.dump.v1`)

## Harness Behavior

The harness executes a deterministic simulation (`ami/runtime/tester`). It copies
`input` to `output` by default. Reserved inputs allow controlled behavior:
- `sleep_ms`: delay execution by the given milliseconds (for timeout tests)
- `error_code`: force an error with the given code string

## Example

```
#pragma test:case c1
#pragma test:runtime input={"x":1} output={"x":1}
#pragma test:kv ns="P1/N1" put="a=1;b=2" get="a" emit=true
```

Running `ami test` will:
- Execute `c1` with the provided input and validate output
- Perform KV puts in namespace `P1/N1` before execution and a get after
- Write per-case metrics and dump JSON under `build/test/kv/`

## Notes

- KV store supports namespacing, TTL (sliding), delete-on-read, and capacity with LRU eviction.
- `ami test --kv-metrics` and `--kv-dump` emit process-level KV artifacts under `build/test/kv/`.
- See `src/ami/runtime/kvstore` for API details; metrics fields: `hits`, `misses`, `expirations`, `evictions`, `currentSize`.

