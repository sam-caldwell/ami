package llvm

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCompileLLToObject_Fail_CapturesStderr(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found") }
    dir := filepath.Join("build", "test", "llvm_compile_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Write invalid LLVM IR to force clang to emit an error
    ll := `; ModuleID = "bad"
target triple = "arm64-apple-macosx"
define i32 @main() {
entry:
  %x = add i32 1, %y ; use of undefined %y should error
  ret i32 %x
}`
    llp := filepath.Join(dir, "bad.ll")
    if err := os.WriteFile(llp, []byte(ll), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := CompileLLToObject(clang, llp, filepath.Join(dir, "bad.o"), DefaultTriple); err == nil {
        t.Fatalf("expected clang failure for invalid IR")
    } else if te, ok := err.(ToolError); !ok || te.Stderr == "" {
        t.Fatalf("expected ToolError with stderr; got %T: %v", err, err)
    }
}

