# Concurrency Guide (Multi‑Agent Development)

Short rules to keep two or more Codex CLI instances productive in this repo.

## Goals
- Minimize merge conflicts and rework.
- Keep outputs deterministic; avoid thrashing shared state.
- Preserve public APIs while landing incremental features.

## Core Guardrails
- One declaration per file: follow the token package pattern project‑wide.
- Narrow blast radius: stay within the active feature/package; no drive‑by refactors.
- Separate mechanical from semantic changes; land mechanics first with tests.
- Deterministic outputs: stable JSON ordering, ISO‑8601 UTC timestamps; no prompts; errors to stderr.

## Command Wiring Pattern
- Pre‑scaffold all subcommands early so `root.go` stays stable.
  - Each subcommand lives in its own file under `src/cmd/ami/cmd/` and exposes `New<Name>Cmd()`.
  - Root imports and registers all known subcommands from day one (stubs allowed).
  - Avoid `init()` side‑effects for command registration (keeps packages declarative).
- Adding features should edit only the subcommand’s file, not `root.go`.

## Tests & Artifacts
- Tests write under `build/test/<pkg>/...` using deterministic file names.
- Each test cleans up its own artifacts; avoid calling `ami clean` in parallel with other agents’ tests.
- Use golden files where helpful; normalize timestamps in assertions.

## Docs & Spec Edits
- SPECIFICATION.md remains the checklist of record; keep edits small and append‑only.
- Roadmap: docs/roadmap.md is the sequencing guide; update when priorities change.
- To avoid collisions, add per‑feature notes in `docs/notes/<feature>.md` and link from SPECIFICATION.md if needed.

## Dependencies & Tooling
- `go.mod` changes have a single owner per sprint; batch version bumps.
- Avoid unsynchronized `go get` across agents; coordinate before adding deps.
- Do not introduce package‑level side effects; keep leaf packages declarative.

## File/Package Ownership
- Claim work at the package or subcommand level.
- Do not rename/move files across packages during another agent’s work.
- Follow naming by concept (e.g., `position.go`, `kind.go`, `sumcheck.go`).

## Conflict Resolution
- If a hotspot is unavoidable (e.g., shared registry), timebox a short ownership window to land changes.
- Prefer adding files over editing existing ones; when editing, keep diffs minimal and focused.

## Pre‑Push Checklist
- `go build -o build/ami ./src/cmd/ami` succeeds.
- `go vet ./...` and `go test -v ./...` pass; ≥80% coverage for changed packages.
- Outputs deterministic; JSON vs human rendering validated where relevant.
- No cross‑package churn or side‑effects added.

## Parallelizable Examples (Early)
- Agent A: `workspace` + `ami init` (and tests).
- Agent B: logging/diagnostics skeleton + flag validation + `ami clean`.
- Next: `lint` (Stage A) vs `mod clean/list`; `token/scanner` vs `source/fileset` in compiler.

## Caution with Clean
- `ami clean` removes and recreates `./build`; announce before running during active work.

— Keep changes small, isolated, and well‑tested. This enables safe concurrency across agents.
## Runtime Scheduler (Developer Notes)
- Worker model per node kind with policies: `fifo`, `lifo`, `fair`, `worksteal`.
- Limits: configure `workers>=1`; construction fails fast on invalid values.
- Backpressure: FIFO/LIFO buffers with `block`, `dropOldest`, `dropNewest`, `shuntOldest`, `shuntNewest`.
- Merge operator: honors `Buffer/Stable/Sort/Key/PartitionBy/Dedup`, optional `Watermark` and `Timeout`, and `Window` bounds.
- Packages:
  - `src/ami/runtime/scheduler`: worker pool with policies and per-kind instantiation.
  - `src/ami/runtime/buffer`: FIFO/LIFO with counters and policies.
  - `src/ami/runtime/merge`: plan + operator with deterministic ordering.
