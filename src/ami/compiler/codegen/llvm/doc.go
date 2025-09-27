package llvm

// Package llvm contains a minimal textual LLVM IR emitter for the AMI IR.
// It lowers a conservative subset sufficient for early backend scaffolding
// and golden tests. Emission focuses on determinism and avoids exposing raw
// pointers at the AMI surface; internal LLVM constructs may use SSA values.

