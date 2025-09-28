package llvm

import (
    "os"
    "path/filepath"
    "runtime"
    "testing"
)

func TestLinkObjects_Fail_CapturesStderr(t *testing.T) {
    clang, err := FindClang()
    if err != nil { t.Skip("clang not found") }
    dir := filepath.Join("build", "test", "llvm_link_fail")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    bin := filepath.Join(dir, "app")
    if runtime.GOOS == "windows" { bin += ".exe" }
    // Pass a non-existent object path to force link failure
    if err := LinkObjects(clang, []string{filepath.Join(dir, "missing.o")}, bin, DefaultTriple); err == nil {
        t.Fatalf("expected link failure")
    } else if te, ok := err.(ToolError); !ok || te.Stderr == "" {
        t.Fatalf("expected ToolError with stderr; got %T: %v", err, err)
    }
}

