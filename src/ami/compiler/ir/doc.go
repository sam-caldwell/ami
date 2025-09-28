// Package ir defines the intermediate representation used by AMI.
//
// Design
//
// The IR is a minimal, SSA-like representation with typed values, instructions,
// and basic blocks grouped into functions. Modules carry package-level
// directives and configuration relevant to later lowering passes (e.g.,
// concurrency, backpressure, telemetry, capabilities, and trust level).
//
// Invariants
//
//   - Determinism: IR encoding uses stable key ordering and deterministic
//     traversal. Debug artifacts (`ir.v1`) are sorted to remain stable across
//     runs given identical inputs.
//   - ABI safety: No raw pointers are part of the language-level ABI. Values
//     for containers/events map to opaque handles in the backend. Pointer-like
//     values must not escape public signatures.
//   - SSA discipline: temporaries produced by expressions are immutable and do
//     not alias; mutation is modeled explicitly by assignment ops.
//   - Blocks: straight-line instruction sequences; control-flow ops (loop/goto)
//     are represented in a structured, portable form for lowering.
//
// Surface
//
//   - Module: package name, functions, directives, config.
//   - Function: name, params/results, blocks.
//   - Instructions: VAR, ASSIGN, RETURN, DEFER, EXPR (call/bin/unary), LOOP,
//     GOTO/SETPC/DISPATCH (scaffold for structured control lowering).
//   - Values: carry an ID and a textual type name which the backend maps to
//     target types conservatively.
//
// Encoding
//
// The Encode helpers produce `ir.v1` JSON suitable for debug artifacts under
// `build/debug/ir/<pkg>/<unit>.ir.json`. These are considered implementation
// details for inspection and tests rather than a stable external format.
//
// Backends
//
// Backends consume this IR to produce textual LLVM IR and objects. They must
// preserve the above invariants and avoid exposing raw pointers in public ABI.
package ir
