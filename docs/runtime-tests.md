# AMI Runtime Tests (Phase 2)

Goal: provide an executable runtime harness for AMI `*_test.ami` cases with deterministic scheduling, explicit fixtures, and structured JSON output.

## Case Shape (proposal)

- `#pragma test:case <name>`: Defines a logical runtime test case.
- `#pragma test:runtime pipeline=<name> input=<json> expect_output=<json> [timeout=<ms>]`
  - Executes the given pipeline with the provided JSON payload and asserts an exact JSON match on output.
- `#pragma test:runtime expect_error=<CODE> [msg~="substr"] [timeout=<ms>]`
  - Executes the pipeline and asserts a specific runtime error.
- `#pragma test:fixture path=<rel> mode=<ro|rw>`: Opts into limited file fixtures.

## Determinism

- No network or external I/O unless explicitly permitted via fixtures.
- Bounded buffers, explicit backpressure policies, and stable scheduling.
- Per‑case timeouts and reproducible ordering.

## CLI Integration

- `ami test` will parse these pragmas from `*_test.ami` and delegate to a runtime tester once pipelines are compilable.
- JSON output conforms to `test.v1` with `test_start/test_end` events and attached `diag.v1` diagnostics on failure.

## Current Harness Behavior (Phase 2 scaffold)

- Unknown pipelines behave as identity functions over the input payload.
- Reserved input keys interpreted by the harness:
  - `sleep_ms`: integer delay before producing output (enables timeout tests).
  - `error_code`: string to simulate a runtime error with that code.
- Assertions:
  - When `expect_error` is set, the simulated error must match to pass; otherwise the case fails.
  - When `expect_output` is set, the produced output (input with meta keys removed) must match exactly.
- Timeouts: when `timeout=<ms>` is set and `sleep_ms` exceeds it, the case times out; expecting `E_TIMEOUT` will pass that case.
- Fixtures: `test:fixture path=<rel> mode=<ro|rw>` pragmas are parsed and validated (mode), and are attached to cases. Enforcement of file access is deferred to a later phase.
