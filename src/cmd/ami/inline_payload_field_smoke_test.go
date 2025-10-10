package main

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
    "io"
    llvme "github.com/sam-caldwell/ami/src/ami/compiler/codegen/llvm"
)

// Build the inline-body-demo example and assert runtime.ll includes field helper calls.
func TestInlineBodyDemo_FieldHelpers_InRuntimeLL(t *testing.T) {
    if _, err := llvme.FindClang(); err != nil { t.Skip("clang not found") }
    example := filepath.Join("..", "..", "..", "examples", "inline-body-demo")
    tmp := filepath.Join("build", "test", "inline_body_demo_smoke")
    _ = os.RemoveAll(tmp)
    if err := copyDir(example, tmp); err != nil { t.Fatalf("copyDir: %v", err) }
    // build the copied workspace from a stable path
    if err := runBuild(os.Stdout, tmp, false, true); err != nil {
        t.Fatalf("runBuild inline-body-demo: %v", err)
    }
    // Look for IR JSON in debug and check the op is present
    found := false
    root := filepath.Join(tmp, "build", "debug", "ir")
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() { return nil }
        if strings.HasSuffix(path, ".ir.json") {
            b, _ := os.ReadFile(path)
            if strings.Contains(string(b), "event.payload.field.") { found = true; return io.EOF }
        }
        return nil
    })
    if !found {
        t.Fatalf("did not find event.payload.field in any IR JSON under %s", root)
    }
}
