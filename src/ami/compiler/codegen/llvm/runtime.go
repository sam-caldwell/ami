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
    // zeroization helper: overwrites n bytes at p with 0x00 deterministically
    s += "define void @ami_rt_zeroize(ptr %p, i64 %n) {\n" +
        "entry:\n  br label %loop\n" +
        "loop:\n  %i = phi i64 [ 0, %entry ], [ %next, %loop ]\n  %done = icmp uge i64 %i, %n\n  br i1 %done, label %exit, label %body\n" +
        "body:\n  %addr = getelementptr i8, ptr %p, i64 %i\n  store i8 0, ptr %addr, align 1\n  %next = add i64 %i, 1\n  br label %loop\n" +
        "exit:\n  ret void\n}\n\n"
    // Owned ABI using fixed-size side table (bring-up)
    // Owned ABI using heap-allocated handle { i8* data; i64 len }
    s += "declare ptr @malloc(i64)\n\n"

    s += "define ptr @ami_rt_owned_new(i8* %data, i64 %len) {\n" +
        "entry:\n  %mem = call ptr @malloc(i64 16)\n  %pfield = bitcast ptr %mem to ptr\n  store ptr %data, ptr %pfield, align 8\n  %lenptr.i8 = getelementptr i8, ptr %mem, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  store i64 %len, ptr %lfield, align 8\n  ret ptr %mem\n}\n\n"

    s += "define i64 @ami_rt_owned_len(ptr %h) {\n" +
        "entry:\n  %lenptr.i8 = getelementptr i8, ptr %h, i64 8\n  %lfield = bitcast ptr %lenptr.i8 to ptr\n  %l = load i64, ptr %lfield, align 8\n  ret i64 %l\n}\n\n"

    s += "define ptr @ami_rt_owned_ptr(ptr %h) {\n" +
        "entry:\n  %pfield = bitcast ptr %h to ptr\n  %p = load ptr, ptr %pfield, align 8\n  ret ptr %p\n}\n\n"

    s += "define void @ami_rt_zeroize_owned(ptr %h) {\n" +
        "entry:\n  %p = call ptr @ami_rt_owned_ptr(ptr %h)\n  %n = call i64 @ami_rt_owned_len(ptr %h)\n  call void @ami_rt_zeroize(ptr %p, i64 %n)\n  ret void\n}\n\n"
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
