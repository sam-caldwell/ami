# Concurrency Declarations (Imperative Subset 6.4 / 2.3.6)

AMI supports top-level concurrency and scheduling hints via compiler pragmas.

Pragmas

- `#pragma concurrency <N>`: Sets worker pool size for the unit. `N` must be a positive integer.
- `#pragma scheduling <policy>`: Sets a scheduling hint for the runtime. Policy is a free-form token in this 
  scaffold (e.g., `fair`, `batch`).

IR Exposure

- The compiler records these as module attributes. Assembly listings include header comments 
  like `; concurrency 8` and `; scheduling fair`.

Tests

- See `src/ami/compiler/driver/driver_concurrency_test.go` for end-to-end tests ensuring these pragmas flow into 
  the generated assembly.

Notes

- Workspace `toolchain.compiler.concurrency` remains the authoritative build-time default. The pragma provides a 
  per-unit override and is captured for debug artifacts in this scaffold.

