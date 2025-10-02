// Package sem performs semantic analysis over the AST.
//
// Role
//
//   - Name resolution, scoping, and type checking for declarations, statements,
//     and expressions.
//   - Enforcement of language rules with stable diagnostics; errors aim to be
//     precise and actionable.
//   - Prepares data needed for IR construction and later code generation while
//     keeping the frontend deterministic and hermetic.
//
// Boundaries
//
//   - Does not perform target lowering; that responsibility belongs to the IR
//     and codegen packages.
//   - Avoids side effects and I/O; analysis is a pure transformation from AST
//     to annotated structures and diagnostics.
//
// Invariants
//
//   - Deterministic analysis for identical inputs (AST + package config).
//   - Diagnostic codes and messages are stable to support tooling and tests.
package sem
