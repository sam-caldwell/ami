package llvm

import (
    "os"
    "path/filepath"
)

// RuntimeLL returns a minimal LLVM IR module string providing runtime symbols
// required by generated code and, optionally, a trivial entrypoint `main`.
// The module sets the provided target triple when non-empty; otherwise uses DefaultTriple.
func RuntimeLL(triple string, withMain bool) string {
    if triple == "" { triple = DefaultTriple }
    // Keep output deterministic and minimal.
    // Provide no-op implementations for a small set of runtime functions used by scaffolding.
    // main returns 0 to allow linking an executable during early bring-up.
    s := "; ModuleID = \"ami:runtime\"\n" +
        "target triple = \"" + triple + "\"\n\n" +
        "; minimal runtime stubs for bring-up\n" +
        "define void @ami_rt_panic(i32 %code) {\n" +
        "entry:\n  ret void\n}\n\n" +
        "define ptr @ami_rt_alloc(i64 %size) {\n" +
        "entry:\n  ret ptr null\n}\n\n"
    if withMain {
        s += "define i32 @main() {\nentry:\n  ret i32 0\n}\n"
    }
    return s
}

// WriteRuntimeLL writes the runtime LLVM IR text to the given directory and returns the file path.
func WriteRuntimeLL(dir, triple string, withMain bool) (string, error) {
    if triple == "" { triple = DefaultTriple }
    if err := os.MkdirAll(dir, 0o755); err != nil { return "", err }
    path := filepath.Join(dir, "runtime.ll")
    return path, os.WriteFile(path, []byte(RuntimeLL(triple, withMain)), 0o644)
}

