package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensures binary comparison lowers to an EXPR with op "eq" in IR debug JSON.
func TestDriver_Lower_Compare_Expr_EQ(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ 1 == 2 }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns, _ := obj["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions in IR: %v", obj) }
    // Walk instructions to find an EXPR with op "eq"
    found := false
    for _, f := range fns {
        fn, _ := f.(map[string]any)
        blks, _ := fn["blocks"].([]any)
        for _, bb := range blks {
            bbo, _ := bb.(map[string]any)
            instrs, _ := bbo["instr"].([]any)
            for _, in := range instrs {
                mo, _ := in.(map[string]any)
                if mo["op"] == "EXPR" {
                    if mo["expr"].(map[string]any)["op"] == "eq" { found = true }
                }
            }
        }
    }
    if !found { t.Fatalf("missing eq expr in IR: %s", string(b)) }
}

