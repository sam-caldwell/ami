AMI — Asynchronous Machine Interface
===================================

AMI is a programming language and toolchain built around Pipeline‑Oriented Programming (POP). POP models programs as pipelines of steps connected by typed events. Each pipeline has a clear ingress/egress, deterministic edges, and explicit concurrency and backpressure policies. The compiler enforces safety properties (pointer‑free public ABI, RAII) and produces deterministic debug artifacts alongside per‑target binaries.

Why POP?
- Deterministic composition: pipelines are explicit graphs, not implicit control flow.
- Observability by design: the compiler emits normalized AST/IR/graph artifacts for inspection and testing.
- Strong safety: ABI rules forbid raw pointers at language boundaries; RAII analysis catches leaks and misuse early.

Quick Start
1) Build the CLI
   - `go build -o build/ami ./src/cmd/ami`
2) Run tests
   - `go test ./...`
   - CI enforces a coverage gate (see `docs/toolchain/ci.md`).
3) Try examples
   - `make examples` or see `docs/toolchain/examples.md` and `examples/` for a quick POP walkthrough.

Documentation Index (docs/*)
- Language and Semantics
  - `docs/language/events.md` and `docs/language/language-events.md` — event model and contracts
  - `docs/language/language-workers.md` — workers and execution model
  - `docs/language/language-decorators.md` — function decorators
  - `docs/language/language-enum.md` — enums: naming, determinism, JSON/text
  - `docs/language/grammar-ebnf.md` — EBNF overview for the language
- Pipelines and Concurrency
  - `docs/language/edges.md` — edges, backpressure, and type validation
  - Runtime honors edge policies: edges.json now includes `minCapacity`, `maxCapacity`, and `backpressure` (block, dropOldest, dropNewest). The executor sizes Go channels from `maxCapacity` and emulates backpressure policies per edge.
  - `docs/toolchain/pipelines-v1-quickstart.md` — reading pipelines.v1 (debug) outputs
  - `docs/toolchain/scheduler-guide.md` — scheduling and worker configuration
  - `docs/toolchain/concurrency.md` — concurrency notes and invariants
- Compiler and Runtime
  - `docs/toolchain/compiler/architecture.md` — compiler architecture and phases
  - `docs/toolchain/ir-indices.md` — IR indices and determinism rules
  - `docs/toolchain/runtime-kvstore.md` — runtime key/value store
  - `docs/toolchain/runtime-tests.md` — runtime test patterns and harness
  - `docs/language/signal.md` — planned AMI example for stdlib signal
- Diagnostics and CI
  - `docs/diag-codes.md` — generated diagnostic codes and data keys (with examples)
  - `docs/toolchain/merge-field-diagnostics.md` — merge/field diagnostics and patterns
  - `docs/toolchain/ci.md` — CI coverage gate and configuration
  - `docs/gaps.md` — spec/documentation reconciliation notes
- Workspace and Tooling
  - `docs/toolchain/workspace-audit.md` — workspace structure, dependency audit
  - `docs/toolchain/make.md` — Makefile targets (build, test, bench, e2e)
  - `docs/toolchain/examples.md` — building and exploring the examples
  - `docs/language/stdlib/os.md` — stdlib os package (process runner)
  - `docs/toolchain/examples/os-exec.md` — planned AMI example using stdlib os once hooks are available
  - `docs/language/stdlib/time.md` — stdlib time package (sleep, now, arithmetic, ticker)
  - `docs/language/time.md` — planned AMI example for stdlib time

Test Timeouts and Slower Builders
- Many end-to-end and LLVM integration tests run external tools (e.g., `go`, `git`, `clang`) and are wrapped with context timeouts to avoid hangs.
- You can scale these timeouts via the environment variable `AMI_TEST_TIMEOUT_SCALE`.
  - Example: `AMI_TEST_TIMEOUT_SCALE=2 go test ./...` doubles timeouts.
  - Example: `AMI_TEST_TIMEOUT_SCALE=0.5 go test ./...` halves timeouts.
- This scaling is applied through `src/testutil/timeout.go` and used across e2e and integration tests.

Authoritative Specification
- The authoritative specification is maintained as a `.docx` and mirrored by the YAML tracker:
  - `docs/Asynchronous Machine Interface.docx`
  - `work_tracker/specification-v.0.0.1.yaml`

Project Goals (high level)
- Deterministic builds and artifacts across platforms
- Statically enforced safety constraints (no raw pointers at ABI boundaries; RAII checks)
- First‑class diagnostics for generics, calls/returns, and pipelines
- Predictable, reproducible development and CI workflows

Contributing
- Run `go vet ./...` and `go test ./...` locally.
- The CI coverage gate enforces ≥0.80 coverage for changed packages; see `docs/toolchain/ci.md`.
- To regenerate diagnostics reference after changes, run: `make gen-diag-codes`.
- Examples: `make examples` builds and stages all example workspaces under `build/examples/`.

License
This project is licensed under the terms of the file `LICENSE.txt`.
