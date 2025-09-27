package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Ensure that when signatures have real names but use 'any' types, the IR sig.params
// still emits deterministic name/type objects with type 'any', not inferred arg types.
func TestCompile_IR_ParamNames_AnyTypes_ArePreserved(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc Callee(a any, b any) {}\nfunc F(){ Callee(1, \"x\") }\n"
    fs.AddFile("anynames.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    if len(diags) != 0 { t.Fatalf("unexpected diags: %+v", diags) }
    path := filepath.Join("build", "debug", "ir", "app", "anynames.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns := obj["functions"].([]any)
    found := false
    for _, fnv := range fns {
        fn := fnv.(map[string]any)
        blks := fn["blocks"].([]any)
        for _, bv := range blks {
            blk := bv.(map[string]any)
            for _, iv := range blk["instrs"].([]any) {
                in := iv.(map[string]any)
                if in["op"] != "EXPR" { continue }
                expr := in["expr"].(map[string]any)
                if expr["op"] == "call" && expr["callee"] == "Callee" {
                    sig := expr["sig"].(map[string]any)
                    ps := sig["params"].([]any)
                    if len(ps) != 2 { t.Fatalf("params len=%d", len(ps)) }
                    p0 := ps[0].(map[string]any)
                    p1 := ps[1].(map[string]any)
                    if p0["name"] != "a" || p1["name"] != "b" { t.Fatalf("expected real names a,b; got %v,%v", p0["name"], p1["name"]) }
                    if p0["type"] != "any" || p1["type"] != "any" { t.Fatalf("expected 'any' types preserved; got %v,%v", p0["type"], p1["type"]) }
                    found = true
                }
            }
        }
    }
    if !found { t.Fatalf("Callee call with any-typed param names not found in IR") }
}

