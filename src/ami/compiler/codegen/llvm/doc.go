// Package llvm emits textual LLVM IR for AMI's intermediate representation.
//
// Purpose
//
//   - Provide a lightweight, deterministic emitter suitable for early backend
//     scaffolding and golden tests.
//   - Map ir.Module/ir.Function/Value forms to a conservative subset of LLVM
//     IR while preserving ABI safety guarantees.
//
// Design
//
//   - Deterministic emission: stable ordering of modules, functions, blocks,
//     and instructions to enable reproducible outputs.
//   - ABI safety: avoid exposing raw pointers in public signatures; use opaque
//     handles where needed and keep pointer-like values internal to LLVM.
//   - Minimal dependencies: keep the emitter small and predictable to ease
//     maintenance and debugging.
package llvm
