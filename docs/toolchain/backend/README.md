# Backend: IR → LLVM → Objects

This document explains the current backend design and how the compiler lowers
front‑end IR to textual LLVM IR, then to objects and linked binaries. It also
captures determinism rules and how to reproduce a build locally.

## Goals

- Deterministic textual LLVM IR emission for inspection and testing
- Stable target triple mapping from `os/arch` environments
- Opaque ABI for pointer‑like values at the language boundary (no raw pointers)
- Optional object compilation and linking when LLVM toolchain is available

## Architecture

- IR is produced by the compiler front‑end and encoded to JSON for debug under
  `build/debug/ir/<pkg>/<unit>.ir.json`.
- The LLVM emitter (`src/ami/compiler/codegen/llvm`) converts IR modules into
  textual LLVM IR with a header containing a `target triple`.
- When `--verbose`, `.ll` files are written to `build/debug/llvm/<pkg>/<unit>.ll`.
- If the host has a working LLVM toolchain (`clang`), `.ll` may be compiled to
  `.o` under `build/obj/<pkg>/<unit>.o` and summarized via `objindex.v1`.

## Target Triples

Triples are derived from `toolchain.compiler.env` entries (`os/arch`). Unknown
values default to `arm64-apple-macosx`.

- `darwin/arm64` → `arm64-apple-macosx`
- `darwin/amd64` → `x86_64-apple-macosx`
- `linux/arm64`  → `aarch64-unknown-linux-gnu`
- `linux/amd64`  → `x86_64-unknown-linux-gnu`
- `windows/arm64` → `aarch64-pc-windows-msvc`

See `src/ami/compiler/codegen/llvm/target.go` for the full mapping and tests.

## ABI and Type Mapping

Public ABI avoids exposing raw pointers. The emitter maps AMI surface types to
LLVM types conservatively:

- Scalars: `bool→i1`, `int→i64`, `uint→i64`, `float64|real→double`
- Strings and container/generic types (`Event<T>`, `Owned<T>`, `map`, `set`,
  `slice`, etc.) → `ptr` (opaque handles)
- Public function signatures use `abiType()` which maps any pointer‑like type to
  `i64` handles to avoid raw pointer exposure at the language boundary.

See `types.go` and `abi.go` for details.

## Lowering Coverage

- Functions: single‑result (scaffold); multi‑result deferred
- Instrs: `VAR`, `ASSIGN`, `RETURN`, `EXPR(call, bin, unary)`, bitwise ops, `not`
- Comparisons: int/double forms (`icmp`, `fcmp`) with deterministic mnemonics
- Calls: runtime calls keep `ptr` returns; user calls use ABI‑safe mapping

Unsupported or unknown instructions are emitted as deterministic comments to
preserve artifact stability during development.

## Diagnostics

- `E_LLVM_EMIT` is reported when textual emission fails during `ami build`.
- Toolchain issues are surfaced as `E_TOOLCHAIN_MISSING` or `E_OBJ_COMPILE_FAIL`
  with best‑effort stderr capture.

## Determinism

- Stable header `; ModuleID = "ami:<pkg>/<unit>"` and `target triple = ...`
- Functions and extern declarations are appended in a deterministic order.
- Debug artifacts (AST/IR/LLVM/ASM) are only written in `--verbose` mode.

## Reproducing a Build

1. `make clean && make build`
2. `ami build --verbose`
3. Inspect artifacts under `build/debug/ir/` and `build/debug/llvm/`
4. If `clang` is available, `.o` objects will appear under `build/obj/<pkg>/`.

## Tests

Emitter and backend behavior are covered by unit tests under
`src/ami/compiler/codegen/llvm/*_test.go` and integration tests in
`src/ami/compiler/driver/*_test.go` and `src/cmd/ami/*_test.go`. Tests validate
triples, basic lowering, debug emission, and guarded object compilation.
