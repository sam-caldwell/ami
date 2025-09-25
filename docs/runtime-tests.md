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

## Scaffold

The package `ami/runtime/tester` ships a `Runner` stub that currently marks all cases as `skip` with reason "runtime disabled". This will be replaced with a real executor as the codegen/runtime mature.

