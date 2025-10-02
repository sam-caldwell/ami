// Package types defines the AMI type system representations.
//
// Scope
//
//   - Primitive and composite kinds, function signatures, and type constructors.
//   - Assignability, equivalence, and conversion rules used by semantic
//     analysis and IR formation.
//   - Deterministic string forms for diagnostics and debug artifacts.
//
// Design
//
// The package is deliberately self-contained to avoid cross-package churn.
// Type values favor value semantics; where references are used they should not
// leak into public ABI surfaces. String representations are stable and intended
// for tests and tooling.
package types
