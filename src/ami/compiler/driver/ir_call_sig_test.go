package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestCompile_IR_CallIncludesSignatureBlock(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc Callee(a string, b int) (int) { return b }\nfunc F(){ Callee(\"x\", 1) }\n"
    fs.AddFile("sig.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    if len(diags) != 0 { t.Fatalf("unexpected diags: %+v", diags) }
    path := filepath.Join("build", "debug", "ir", "app", "sig.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    var found bool
    for _, fnv := range fns {
        fn := fnv.(map[string]any)
        blks := fn["blocks"].([]any)
        for _, bv := range blks {
            blk := bv.(map[string]any)
            for _, iv := range blk["instrs"].([]any) {
                in := iv.(map[string]any)
                if in["op"] == "EXPR" {
                    expr := in["expr"].(map[string]any)
                    if expr["op"] == "call" && expr["callee"] == "Callee" {
                        sig := expr["sig"].(map[string]any)
                        ps := sig["params"].([]any)
                        rs := sig["results"].([]any)
                        if len(ps) == 2 && ps[0] == "string" && ps[1] == "int" && len(rs) == 1 && rs[0] == "int" {
                            found = true
                        }
                    }
                }
            }
        }
    }
    if !found { t.Fatalf("call signature not found in IR") }
}

