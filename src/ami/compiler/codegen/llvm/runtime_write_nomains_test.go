package llvm

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// TestWriteRuntimeLL_NoMain writes runtime without main and verifies file exists and contains expected header.
func TestWriteRuntimeLL_NoMain(t *testing.T) {
    dir := filepath.Join("build", "test", "runtime_nomains")
    _ = os.RemoveAll(dir)
    path, err := WriteRuntimeLL(dir, DefaultTriple, false)
    if err != nil { t.Fatalf("write: %v", err) }
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read: %v", err) }
    s := string(b)
    if !strings.Contains(s, "ModuleID = \"ami:runtime\"") {
        t.Fatalf("missing runtime header: %s", s)
    }
    if strings.Contains(s, "define i32 @main()") {
        t.Fatalf("unexpected main() present in runtime without main flag")
    }
}
