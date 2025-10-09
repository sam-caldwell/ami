package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure gpu(...) blocks are lowered into IR.Function.GPUBlocks and encoded in IR JSON.
func TestDriver_GPUBlock_Lowering_And_IR_Encode(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ gpu(family=\"metal\", name=\"k\", n=4){ kernel void k(){} } }\n"
    fs.AddFile("unit1.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    if len(diags) != 0 {
        b, _ := json.Marshal(diags)
        t.Fatalf("unexpected diagnostics: %s", string(b))
    }
    // Read IR JSON emitted for the unit
    irPath := filepath.Join("build", "debug", "ir", "app", "unit1.ir.json")
    b, err := os.ReadFile(irPath)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fl, _ := obj["functions"].([]any)
    if len(fl) == 0 { t.Fatalf("no functions in IR") }
    fn := fl[0].(map[string]any)
    gbs, _ := fn["gpuBlocks"].([]any)
    if len(gbs) == 0 { t.Fatalf("expected gpuBlocks in IR") }
    gb := gbs[0].(map[string]any)
    if gb["family"] != "metal" { t.Fatalf("family: %v", gb["family"]) }
    if gb["name"] != "k" { t.Fatalf("name: %v", gb["name"]) }
    if v, ok := gb["n"].(float64); !ok || int(v) != 4 { t.Fatalf("n: %v", gb["n"]) }
    if s, _ := gb["source"].(string); s == "" { t.Fatalf("source missing") }
}
