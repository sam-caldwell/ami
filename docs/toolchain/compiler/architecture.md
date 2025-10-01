# AMI Compiler Architecture

This document summarizes the architecture of the AMI compiler as implemented in
this repository. The goal is a deterministic pipeline from sources (`.ami`) to
IR to backend codegen and artifacts, with clear seams for testing and evolution.

## Packages & Responsibilities

- `src/ami/compiler/source`: FileSets and precise positions for `.ami` sources.
- `src/ami/compiler/token` / `scanner` / `parser`: Lex/parse to AST with
  tolerant recovery and position‑rich diagnostics.
- `src/ami/compiler/ast`: Nodes for files/decls/stmts/exprs/types; pragmas.
- `src/ami/compiler/types`: Primitive and composite type renderers.
- `src/ami/compiler/sem`: Symbol tables/scopes; memory safety checks; pipeline
  invariants; RAII analysis; type inference passes (M1–M3).
- `src/ami/compiler/ir`: SSA‑like IR: ops, typed values, functions, blocks.
- `src/ami/compiler/driver`: Orchestration for lowering AST→IR, debug artifacts
  (`ast.v1`, IR, pipelines, eventmeta, ASM), indices, and deterministic writes.
- `src/ami/compiler/codegen`: Backend registry and LLVM adapter layer.
- `src/ami/compiler/codegen/llvm`: Textual LLVM emission and toolchain bridges
  to `.o` and linking; deterministic target triple mapping.

## Determinism Rules

- Stable ordering for files, functions, externs, debug emissions.
- JSON debug artifacts (`ast.v1`, IR, pipelines, indices) sorted and normalized.
- Emission only in `--verbose` to avoid ambient file churn.

## Diagnostics

- `diag.v1` records with `file` and precise positions where available.
- Memory safety (`E_PTR_UNSUPPORTED_SYNTAX`, `E_MUT_ASSIGN_UNMARKED`,
  `E_MUT_BLOCK_UNSUPPORTED`) from `sem` pass.
- Backend: `E_LLVM_EMIT`, `E_TOOLCHAIN_MISSING`, `E_OBJ_COMPILE_FAIL`, `E_LINK_FAIL`.

## Directives & Config

- Pragmas: `#pragma concurrency`, `#pragma backpressure`, `#pragma telemetry`,
  `#pragma capabilities`, `#pragma trust`.
- Derived attributes (e.g., capabilities from `io.*` steps) included in IR and
  surfaced in debug artifacts.

## Build Command Integration

- `ami build` reads `ami.workspace`, audits dependencies, runs compiler driver,
  writes debug artifacts in verbose mode, compiles per‑env objects and links
  binaries when toolchain present. Build plan/manifests track outputs.

## Testing

- Unit and E2E tests across packages; golden JSON artifacts for debug outputs;
- conditional linking tests when `clang` is available.
