// Package ast defines the abstract syntax tree (AST) for AMI.
//
// Overview
//
// The AST models the surface syntax produced by the parser. Nodes carry source
// positions for diagnostics and are intentionally lightweight to keep parsing
// and semantic analysis fast and deterministic. Desugaring and lowering are
// deferred to later phases (sem and ir/codegen) to keep the AST close to the
// source language.
//
// Conventions
//
//   - Nodes are immutable after construction and are safe to share across
//     passes.
//   - Position fields reference the source package for stable line/column
//     mapping used by diagnostics and tooling.
//   - The package favors clarity over micro-optimizations; helpers for
//     printing/pretty-strings are intended for debugging and tests.
//
// Invariants
//
//   - Trees are acyclic; parent/child relationships are well-formed.
//   - Node spans cover their tokens; diagnostics should reference node spans
//     where possible for consistency.
package ast
