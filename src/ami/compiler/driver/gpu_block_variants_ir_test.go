package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure multiple gpu blocks (metal + opencl) are both encoded in IR.
func TestDriver_GPUBlock_Variants_Metal_And_OpenCL(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ gpu(family=\"metal\", name=\"m\", n=2){ x } gpu(family=\"opencl\", name=\"o\", n=3){ y } }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    if len(diags) != 0 { b, _ := json.Marshal(diags); t.Fatalf("diags: %s", string(b)) }
    irPath := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(irPath)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fl, _ := obj["functions"].([]any)
    if len(fl) == 0 { t.Fatalf("no functions") }
    fn := fl[0].(map[string]any)
    gbs, _ := fn["gpuBlocks"].([]any)
    if len(gbs) != 2 { t.Fatalf("expected 2 gpuBlocks, got %d", len(gbs)) }
}

