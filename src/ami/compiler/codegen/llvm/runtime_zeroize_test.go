package llvm

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// Ensure runtime.ll includes the zeroize helper definition.
func TestRuntimeLL_IncludesZeroize(t *testing.T) {
    dir := filepath.Join("build", "test", "llvm_runtime_zeroize")
    _ = os.RemoveAll(dir)
    ll, err := WriteRuntimeLL(dir, DefaultTriple, false)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read: %v", err) }
    if !strings.Contains(string(b), "define void @ami_rt_zeroize(") {
        t.Fatalf("runtime.ll missing zeroize: %s", string(b))
    }
}

