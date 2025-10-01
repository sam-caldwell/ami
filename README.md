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
   - CI enforces a coverage gate (see `docs/ci.md`).
3) Try examples
   - `make examples` or see `docs/examples.md` and `examples/` for a quick POP walkthrough.

Documentation Index (docs/*)
- Language and Semantics
  - `docs/events.md` and `docs/language-events.md` — event model and contracts
  - `docs/language-workers.md` — workers and execution model
  - `docs/language-decorators.md` — function decorators
  - `docs/language-enum.md` — enums: naming, determinism, JSON/text
  - `docs/grammar-ebnf.md` — EBNF overview for the language
- Pipelines and Concurrency
  - `docs/edges.md` — edges, backpressure, and type validation
  - `docs/pipelines-v1-quickstart.md` — reading pipelines.v1 (debug) outputs
  - `docs/scheduler-guide.md` — scheduling and worker configuration
  - `docs/concurrency.md` — concurrency notes and invariants
- Compiler and Runtime
  - `docs/compiler/architecture.md` — compiler architecture and phases
  - `docs/ir-indices.md` — IR indices and determinism rules
  - `docs/runtime-kvstore.md` — runtime key/value store
  - `docs/runtime-tests.md` — runtime test patterns and harness
  - `docs/language/signal.md` — planned AMI example for stdlib signal
- Diagnostics and CI
  - `docs/diag-codes.md` — generated diagnostic codes and data keys (with examples)
  - `docs/merge-field-diagnostics.md` — merge/field diagnostics and patterns
  - `docs/ci.md` — CI coverage gate and configuration
  - `docs/gaps.md` — spec/documentation reconciliation notes
- Workspace and Tooling
  - `docs/workspace-audit.md` — workspace structure, dependency audit
  - `docs/make.md` — Makefile targets (build, test, bench, e2e)
  - `docs/examples.md` — building and exploring the examples
  - `docs/stdlib/os.md` — stdlib os package (process runner)
  - `docs/examples/os-exec.md` — planned AMI example using stdlib os once hooks are available
  - `docs/stdlib/signal.md` — stdlib signal package (handlers)

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
- The CI coverage gate enforces ≥0.80 coverage for changed packages; see `docs/ci.md`.
- To regenerate diagnostics reference after changes, run: `make gen-diag-codes`.
- Examples: `make examples` builds and stages all example workspaces under `build/examples/`.

License
This project is licensed under the terms of the file `LICENSE.txt`.
