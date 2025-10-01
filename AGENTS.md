# Repository Guidelines

## Project Structure & Module Organization
- `src/` — Go sources: `cmd/ami` (CLI), `ami/compiler` (parser/semantics/IR/codegen), `ami/runtime` (executor/scheduler/merge), `schemas/*`.
- `docs/` — language, runtime, and diagnostics docs.
- `tests/e2e/` — end‑to‑end CLI tests; unit tests live next to code as `*_test.go`.
- `tools/` — utilities (e.g., `gen-diag-codes`). `build/` — generated artifacts.

## Build, Test, and Development Commands
- `make build` — build CLI to `build/ami`.
- `go test ./...` — run all tests. `make test` — verbose variant.
- `make coverage-short` — quick coverage on key packages.
- `make gen-diag-codes` — regenerate `docs/diag-codes.md` from code.
- Run locally: `./build/ami help` or `go run ./src/cmd/ami`.

## Coding Style & Naming Conventions
- Standard Go formatting (`gofmt`/`goimports`); keep diffs minimal.
- Packages lowercase; exported types/functions PascalCase; locals lowerCamel.
- One cohesive declaration per file when practical; colocate tests as `*_test.go`.
- Diagnostic codes use `E_*`/`W_*` (see `docs/diag-codes.md`). Keep outputs deterministic.

## Testing Guidelines
- Use Go `testing` with table‑driven tests; include happy and sad paths.
- Coverage target: ≥0.80 on changed packages; ensure `go vet` and `go test ./...` pass.
- E2E tests reside in `tests/e2e`; unit tests live beside sources.

## Commit & Pull Request Guidelines
- Message style: `area: imperative summary` (e.g., `driver: improve if/else lowering`).
- PRs should describe motivation, link spec/work items, include tests/docs updates, and pass CI.
- Regenerate docs when relevant (e.g., `make gen-diag-codes`).

## Security & Configuration Tips
- Env knobs: `AMI_PACKAGE_CACHE`, `AMI_STRICT_DEDUP_PARTITION`. Gate external tooling (e.g., `AMI_E2E_ENABLE_GIT=1`).
- Do not commit secrets; prefer static, reproducible outputs.

## Architecture Overview (Brief)
- Frontend: scanner → parser → semantics (`sem`), emitting diagnostics.
- IR/codegen: `ir` plus LLVM emission (`codegen/llvm`).
- Runtime: executor/scheduler/merge in `src/ami/runtime`; CLI in `src/cmd/ami`.
