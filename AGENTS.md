# Repository Guidelines

## Project Structure & Module Organization
- `src/` — Go sources: `cmd/ami` (CLI), `ami/compiler` (parser/semantics/IR/codegen), `ami/runtime` (executor/scheduler/merge), `schemas/*`.
- `src/ami/runtime/host/*` — host‑backed stdlib (Go) used by the runtime.
- `std/ami/stdlib/*` — AMI stdlib packages (.ami only; CI guard enforces no Go files).
- `docs/` — language, toolchain, runtime, and diagnostics docs.
- `tests/e2e/` — end‑to‑end CLI tests; unit tests live beside code as `*_test.go`.
- `tools/`, `scripts/`, `build/` — utilities, helper scripts, generated artifacts.

## Build, Test, and Development Commands
- Build CLI: `make build` → `build/ami`.
- Run tests: `go test ./...` (full) or `make coverage-short` (quick CLI focus).
- Coverage gates (≥0.80):
  - Overall: `make coverage-gate-total` (or `bash scripts/coverage_gate_total.sh`).
  - Changed pkgs: `bash scripts/coverage_gate.sh`.
- Run locally: `./build/ami help` or `go run ./src/cmd/ami`.
- Regenerate diag docs: `make gen-diag-codes`.

## Coding Style & Naming Conventions
- Go formatting: `gofmt`/`goimports`; static checks via `go vet` (run by `make lint`).
- One cohesive declaration per `.go` file; colocate tests as `*_test.go`.
- Packages lowercase; exported PascalCase; locals lowerCamel; descriptive filenames (e.g., `io_allowed_ingress.go`).
- Keep diffs minimal and focused; prefer small, surgical changes.

## Testing Guidelines
- Framework: Go `testing`; prefer table‑driven tests with happy/sad paths.
- Coverage: maintain ≥80% across `src/*`; CI enforces overall and changed‑package gates.
- E2E tests: `make e2e-test` or `make e2e-one NAME=Pattern`.
- Test names: `Test<Area>_<Behavior>` (e.g., `TestPipelineSemantics_IO_InMiddle_Error`).

## Commit & Pull Request Guidelines
- Branching: use `main` only.
- Commits: Conventional Commits (e.g., `feat(scanner): add duration literal`) or `area: imperative summary`.
- PRs: clear description (what/why), link spec/work items (e.g., `F-1-4`), include tests/docs, and pass CI (`go vet`, tests, coverage gates).

## Security & Configuration Tips
- I/O gating via `io.Policy` and `exec.SandboxPolicy`; do not bypass in tests.
- Env knobs: `AMI_PACKAGE_CACHE`, `AMI_STRICT_DEDUP_PARTITION`, `AMI_E2E_ENABLE_GIT=1`.
- Stdlib policy: only `.ami` files under `std/ami/stdlib` (CI guard).

## Architecture Overview (Brief)
- Frontend: scanner → parser → semantics.
- IR/Codegen: LLVM emission; math maps to LLVM intrinsics or runtime helpers.
- Runtime: executor/scheduler/merge; host stdlib in `src/ami/runtime/host`.

