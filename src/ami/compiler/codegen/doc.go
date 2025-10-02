// Package codegen handles lowering AMI IR to target artifacts.
//
// Scope
//
// This package provides target-agnostic facades and orchestration used by
// specific backends (e.g., LLVM). It coordinates passes that transform
// ir.Module and ir.Function structures into concrete artifacts while preserving
// determinism and ABI safety constraints.
//
// Outputs
//
//   - Textual IR for inspection and golden tests.
//   - Object files and linkable units when toolchains are available.
//   - Debug artifacts emitted under build/debug when enabled by the driver.
//
// Guarantees
//
//   - Stable emission for identical inputs to support reproducible builds.
//   - No raw pointers are exposed in public ABI boundaries; opaque handles are
//     used where necessary and mapped by the backend.
package codegen
