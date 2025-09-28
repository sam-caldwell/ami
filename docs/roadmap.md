# AMI Toolchain Delivery Roadmap

This roadmap captures the phased, test-first plan to deliver the AMI toolchain and compiler as described in SPECIFICATION.md and the AMI docx. It reflects the agreed priorities and codegen order, and is designed for iterative execution with clear acceptance gates per milestone.

Authoritative source of truth remains `docs/Asynchronous Machine Interface.docx`. SPECIFICATION.md tracks features and checklists. This document guides execution sequencing.

## Decisions & Constraints

- Target matrix: Phase 1 defaults to `darwin/arm64` only; expand later.
- Package fetching scope: local-only early; `git+ssh` later.
- Codegen order: simplified object format → debug ASM → LLVM objects.
- Priority order: bring CLI and linter to green first; then compiler frontend/semantics; then IR/codegen.
- Repository conventions: one declaration per file; ≥80% coverage per touched package; deterministic outputs; JSON/human modes; errors to stderr; JSON lines for streaming.

## Milestones

Each milestone lists key deliverables and acceptance criteria. All milestones must pass `go vet ./...`, `go test -v ./...` (with ≥80% coverage in changed packages), and `go build -o build/ami ./src/cmd/ami`.

### M0 — Bootstrap

Deliverables:
- Directory layout scaffolding under `src/` for: `ami/{exit,logging,workspace}`, `cmd/ami`, `ami/compiler/{source,token,scanner,ast,parser,types,sem,ir,codegen}`.
- Minimal tests to ensure packages compile and run.

Acceptance:
- Build/vet/test green on skeleton; initial coverage recorded.

### M1 — CLI Core + Global Flags

Deliverables:
- Cobra root in `src/cmd/ami/root.go` with persistent flags: `--help`, `--json`, `--verbose`, `--color`.
- Exit codes in `src/ami/exit` with helpers mapping error classes.
- Mutual exclusion validation: `--json` and `--color` rejects with USER_ERROR; error text to stderr.
- Command stubs: `init`, `clean`, `mod {clean,sum,get,list}`, `lint`, `test`, `build`, `version`, `help`.

Acceptance:
- Unit tests cover flag parsing/help text/exit paths; JSON vs human outputs; mutual exclusion behavior.

### M2 — Workspace Management

Deliverables:
- `workspace.Workspace` with `Load/Save/Create` for `ami.workspace`.
- `ami init`: create minimal workspace or `--force` add missing fields; ensure `toolchain.compiler.target` and `packages.main.root` exist; seed `toolchain.compiler.env` with `darwin/arm64`; ensure git repo or `git init` on `--force`; add `.gitignore` with `./build`.

Acceptance:
- Tests: new workspace; idempotent re-init; `--force` with/without missing fields; JSON mode behavior.

### M3 — Logging + Diagnostics

Deliverables:
- `logging` package: JSON renderer (stable schema) and human renderer; verbose lines prefixed with ISO‑8601 UTC (ms). Colors only when human mode and `--color`.
- Diagnostic record schema `diag.v1` with level, code, message, file, and precise positions; JSON lines for streaming.

Follow‑up (M3.1): stdlib logger pipeline configuration
- Expose CLI flags to configure `ami/stdlib/logger` pipeline redaction when the CLI migrates from direct logger to pipeline mode.
- Map root flags to pipeline config: JSONRedactKeys, JSONRedactPrefixes (and future allow/deny once available).
- Add tests for batch/interval/backpressure policies and counters; verify safety‑net redaction for `log.v1` JSON lines.

Acceptance:
- Tests: timestamp format; multi-line handling; color disabled in JSON; fields present; stderr routing for invalid flag combos.

### M4 — Linter (Stage A: Pre‑Parser Checks)

Deliverables:
- `ami lint`: initial checks that do not require full parsing: workspace/package presence; import/version shape validation; basic naming constraints.
- Outputs: human summary and `diag.v1` lines to stdout; final summary record.

Acceptance:
- Tests: happy/sad rule coverage; strict mode filters; JSON/human mode switching.

### M5 — Frontend (Source → AST)

Deliverables:
- `source`: `position.go`, `file.go`, `fileset.go` with precise positions.
- `token`: `kind.go`, `token.go`, `keywords.go`, `symbols.go`.
- `scanner`: UTF‑8, comments, tokens, recovery; position‑rich errors.
- `ast`: nodes for files/decls/stmts/exprs/types; comment attachment.
- `parser`: imports, package decls, funcs, pipelines and `error {}` blocks; tuple returns; container literals; tolerant generics scaffold.
- Verbose `ast.v1` JSON under `build/debug/` with stable ordering.

Acceptance:
- Golden tests for scanner/parser; recovery paths; stable AST JSON.

### M6 — Semantics (M0)

Deliverables:
- `types`: primitives, `Event<T>`, `Error<E>`, `Owned<T>`, `slice<T>`, `set<T>`, `map<K,V>`; renderers.
- `sem`: symbol tables/scopes, intra‑workspace import resolution, const folding, basic assignment/call checks.
- Memory Safety (AMI 2.3.2): ban `&`; unary `*` only LHS mutating marker; diagnostics: `E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_ASSIGN_UNMARKED`, `E_MUT_BLOCK_UNSUPPORTED`.
- Pipelines: structural validation and canonical worker signature constraints (no raw pointers; ambient state per docx rule).

Acceptance:
- Tests for valid/invalid cases with precise positions; coverage ≥80%.

### M7 — Linter (Stage B: Parser‑backed Rules)

Deliverables:
- Integrate parser/semantics: unknown identifiers, unused, import existence/versioning, duplicate imports/aliases, basic formatting markers.
- AMI semantics lints: memory safety; RAII hint `W_RAII_OWNED_HINT`.

Acceptance:
- Tests: rule filtering, strict mode, diagnostics streaming.

### M8 — Type Inference (M1/M2/M3)

Deliverables:
- M1: locals inference; unary/binary ops; call/assign constraints; `E_TYPE_AMBIGUOUS` with positions.
- M2: container element/key inference; tuple return inference; propagation through `Event<T>`/`Error<E>`.
- M3: conservative compatibility for generic `Event<typevar>` across steps; scoping/shadowing; return inference.

Acceptance:
- Targeted unit tests and golden diagnostics; stable positions; coverage maintained.

### M9 — IR + Artifacts

Deliverables:
- IR (SSA-like): ops VAR, ASSIGN, RETURN, DEFER, EXPR; typed values; blocks/edges.
- Lowering: functions, calls, container literals; pipeline edges with capacity/backpressure annotations (scaffold).
- Debug IR JSON under `build/debug/ir/...` with deterministic ordering.
- Non‑debug obj index: `build/obj/<package>/index.json` (`objindex.v1`).

Acceptance:
- Golden IR; schema validation; determinism across runs (normalized timestamps in tests).

### M10 — Codegen & Linking

Deliverables:
- Stage A: simplified object format with relocatable symbol abstraction, indexes, integrity checks.
- Stage B: initial debug ASM emission from IR; stable ASM indexing (for tests and inspection).
- Stage C: LLVM objects generation; relocations; symbol tables.
- Linker: symbol resolution, DCE, relocations, init order, metadata tables; produce PIE/static per config.

Acceptance:
- Minimal and multi‑package builds; missing deps map to correct exits; deterministic outputs.

### M11 — Build Command

Deliverables:
- `ami build`: reads `ami.workspace`, ensures deps available, compiles packages (frontend→sem→IR→codegen), produces artifacts.
- Verbose: build plan JSON and debug artifacts under `build/debug/`.
- Error handling: user vs IO vs integrity; JSON diagnostics streaming.

Acceptance:
- Schema tests for `ast.v1`, IR indices, `objindex.v1`; repeatability checks; failure mode tests.

### M12 — Test Runner

Deliverables:
- `ami test`: collects `_test.go`, produces binary in `build/test`, runs; `--verbose` writes `build/test/test.log` and `build/test/test.manifest`.

Acceptance:
- Harness tests for execution, logs, and manifests; JSON/human mode behavior.

### M13 — Modules/Cache

Deliverables:
- `AMI_PACKAGE_CACHE` detection/creation; default `${HOME}/.ami/pkg`.
- `mod clean`: remove and recreate cache.
- [x] `mod list`: list cached packages.
- `mod sum`: validate `ami.sum`, verify hashes, compare to workspace, update as needed.
- Later: `mod get` with `git+ssh`, SemVer selection, integrity checks (post‑Phase 1).

Acceptance:
- Tests for each subcommand; integrity and error paths; no interactive prompts.

### M14 — RAII + Generic Inference

Dependencies:
- M5 Frontend (Source→AST) complete
- M6 Semantics (initial) complete (memory safety, pipeline invariants, worker signature/resolution)
- M7 Linter Stage B complete (parser-backed rules + pragma suppression)
- M8 Type Inference (M1–M3) baseline in place
- M9 IR + Artifacts scaffold complete (deterministic IR/obj index)
- M12 Test Runner available (directive/runtime harness; JSON lines)

Deliverables:
- Full RAII ownership accounting for `Owned<T>` (intraprocedural, defer-aware):
  - States: acquire/transfer/release; defer-scheduled release counting to function exit
  - Diagnostics: `E_RAII_LEAK`, `E_RAII_DOUBLE_RELEASE`, `E_RAII_USE_AFTER_RELEASE`, `E_RAII_RELEASE_UNOWNED`, `E_RAII_TRANSFER_UNOWNED`
  - Semantics pass `sem.AnalyzeRAII` + position-rich diag.v1
  - Optional IR debug markers for RAII events (acquire/release/defer/transfer)
- Richer generic inference (local, deterministic):
  - Call-site instantiation/unification for single-letter generics
  - Tuple return propagation and arity checks
  - Container inference for `slice<T>`, `set<T>`, `map<K,V>` via literals/assign/return
  - Propagation for `Event<T>` / `Error<E>` across local scopes
  - Diagnostics: `E_TYPE_UNINFERRED`, extended `E_TYPE_MISMATCH`/`E_TYPE_AMBIGUOUS`
- Linter Stage B integration surfaces RAII errors; respects `#pragma lint:disable`.

Acceptance:
- `go vet ./...`, `go test -v ./...` pass; ≥80% coverage in changed packages
- Deterministic diagnostics with precise positions (golden tests for inference/RAII)
- No regressions across earlier milestones; stable JSON/human outputs

### M15 — linux/arm64 Codegen (Cross-Target)

Dependencies:
- M9 IR + Artifacts complete (deterministic IR)
- M10 Codegen & Linking complete for darwin/arm64
- M11 Build Command complete (env matrix, per-env objects/indices, manifest/plan)

Deliverables:
- Code generator supports `linux/arm64` target triple (`aarch64-unknown-linux-gnu`) end-to-end:
  - Emit LLVM IR with `target triple = "aarch64-unknown-linux-gnu"`
  - Compile `.ll` → `.o` for linux/arm64 via clang
  - Link `.o` into linux/arm64 executable with appropriate flags (PIE/static per workspace options)
- `ami build` honors mixed `toolchain.compiler.env` such as `[darwin/arm64, linux/arm64]`:
  - Writes per-env objects under `build/linux/arm64/obj/**`
  - Produces per-env `index.json` and includes paths in build plan and `build/ami.manifest`
  - Attempts per-env linking when the toolchain supports it; logs clear diagnostics when cross-linker is unavailable
- Deterministic build artifacts and diagnostics across both environments.

Acceptance:
- On a host with an appropriate clang toolchain:
  - `ami build` with `toolchain.compiler.env: [linux/arm64]` produces a linked linux/arm64 binary; manifest lists the binary and env-specific objects/indices
  - With `toolchain.compiler.env: [darwin/arm64, linux/arm64]`, per-env objects exist and indexes are recorded; binaries are produced where the host/linker allows; missing cross-linker yields diag with non-fatal logging when permitted by config
- Tests validate per-env object emission and manifest/plan entries for linux/arm64; codegen emits correct triple in LLVM IR.

### M16 — darwin/amd64 and linux/amd64 Codegen (Cross-Target, Generalization)

Dependencies:
- M10 Codegen & Linking complete (baseline)
- M11 Build Command complete (env matrix and per-env artifacts)
- M15 linux/arm64 Codegen complete (pattern established)

Deliverables:
- Add first-class support for `darwin/amd64` and `linux/amd64`:
  - Correct LLVM target triples: `x86_64-apple-macosx` and `x86_64-unknown-linux-gnu`
  - Emit `.ll` with the target triple set per environment
  - Compile `.ll` → `.o` and link binaries via clang for each environment when available
- Ensure code generation is abstract and extensible:
  - Triple selection centralized (`TripleFor(os,arch)`), no special-case branches in lowering
  - Per-env build loop shared for all OS/arch, writing objects under `build/<env>/obj/**`
  - Manifest and build plan include env-specific object indices and objects uniformly
- `ami build` honors env matrices mixing arm64 and amd64 across darwin/linux; links where toolchain permits and emits clear diagnostics otherwise.

Acceptance:
- On hosts with appropriate toolchains, `ami build` for `darwin/amd64` and `linux/amd64` produces linked binaries; manifests list binaries and env objects/indices
- With mixed env matrices, per-env objects exist under `build/<env>/obj/**` with `index.json`; build plan and `build/ami.manifest` capture paths deterministically
- Unit tests verify:
  - Triple mapping for `darwin/amd64` and `linux/amd64`
  - LLVM IR contains the correct `target triple`
  - Per-env object emission and index writing for amd64 envs

## Sprint 1 (Immediate Backlog)

- Scaffold M0/M1: root CLI + flags/exit codes; tests for flag parsing, stderr/stdout behavior.
- Implement `workspace.Workspace` + tests.
- Implement `ami init` and `ami clean` fully with ≥80% coverage.
- Logging/diagnostics skeleton with basic tests.
- Early `ami lint` Stage A rules with tests.
- Keep `go build`, `go vet`, `go test` green.

## Quality Gates & Determinism

- Build consistently with `go build -o build/ami ./src/cmd/ami`.
- `go vet ./...` and `go test -v ./...` pass; ≥80% coverage on changed packages.
- Deterministic outputs: stable JSON key ordering; timestamps are ISO‑8601 UTC with ms and normalized in golden tests.
- Human output to stdout; errors to stderr; JSON mode disables colors and uses NDJSON for streaming.

## Updating This Roadmap

- Treat this roadmap as the sequencing guide; SPECIFICATION.md remains the feature checklist of record.
- When a milestone completes, check it off in SPECIFICATION.md under the relevant feature area and reference commits.
- If priorities change, update the Decisions & Constraints and the affected milestones here; keep changes small and focused.
