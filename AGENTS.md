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
- Branching: DO NOT CHANGE BRANCH.  Use 'main' branch only.
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
# Repository Guidelines

## Project Structure & Module Organization
- `src/ami/...`: Primary Go packages (compiler, runtime, stdlib).
- `src/cmd/ami`: CLI entrypoint for building, linting, and E2E.
- `docs/`: Language, toolchain, and diag docs; `work_tracker/`: specs and IDs (e.g., F‑1‑4).
- `examples/`: Small workspaces for demos; `tests/`: E2E suites; `build/`: artifacts (debug IR/ASM/LLVM, objects).

## Build, Test, and Development Commands
- `make build`: Build the CLI to `build/ami`.
- `make test`: Run all unit tests (`go test -v ./...`).
- `make coverage-short`: Quick CLI coverage + schema sanity.
- `make examples`: Build all example workspaces (stages outputs under `build/examples/`).
- `make e2e-test` / `make e2e-one NAME=Pattern`: Run E2E tests.
- `go test ./... -coverprofile=build/coverage.out`: Full coverage report.

## Coding Style & Naming Conventions
- Go formatting: use `gofmt`/`go vet` (`make lint`). Tabs/standard Go style.
- Keep files small: one top‑level declaration (function/struct/type/method) per `.go` file.
- Naming: package‑scoped files are descriptive and focused (e.g., `io_allowed_ingress.go`, `pipeline_edges_validation_test.go`).
- Prefer additive, surgical changes. Avoid sweeping refactors and cross‑package churn.

## Testing Guidelines
- Framework: Go `testing`. Co‑locate `*_test.go` with code; aim for one focused test per file when practical.
- Coverage: target ≥80% overall; LLVM/codegen paths should remain ≥80%.
- Conventions: `Test<Area>_<Behavior>` (e.g., `TestPipelineSemantics_IO_InMiddle_Error`).
- Utilities: `make test-hotspots` reports packages/files missing paired tests.

## Commit & Pull Request Guidelines
- Commit style: Conventional Commits – `type(scope): summary` (e.g., `feat(scanner): add duration literal scanning`).
- PRs: small diffs, clear description (what/why), linked spec IDs (e.g., `F-1-4`), and validation steps (`make lint && make test`). Include tests and docs for user‑visible changes.
- Avoid unrelated formatting or file moves. Keep changeset localized to the feature/fix.

## Security & Configuration Tips
- I/O gating is enforced via `io.Policy` and runtime `exec.SandboxPolicy`; do not bypass in tests.
- Do not commit secrets; prefer env vars and local configs. Use `make zip` for clean snapshots.

## Agent‑Specific Instructions
- Keep the blast radius small: one declaration per file; prefer new files over large edits.
- Search with `rg`; read files in ≤250‑line chunks; run targeted tests only for changed areas before full suite.
