package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify synthesized param names p0,p1 appear when the callee signature is unknown.
func TestCompile_IR_SynthesizedParamNames_WhenSignatureUnknown(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ External(1,2) }\n"
    fs.AddFile("synth.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, diags := Compile(ws, pkgs, Options{Debug: true})
    if len(diags) != 0 { t.Fatalf("unexpected diags: %+v", diags) }
    path := filepath.Join("build", "debug", "ir", "app", "synth.ir.json")
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
                if expr["op"] == "call" && expr["callee"] == "External" {
                    sig := expr["sig"].(map[string]any)
                    ps := sig["params"].([]any)
                    if len(ps) != 2 { t.Fatalf("params len=%d", len(ps)) }
                    // Expect synthesized names p0 and p1 with int types.
                    p0 := ps[0].(map[string]any)
                    p1 := ps[1].(map[string]any)
                    if p0["name"] != "p0" || p1["name"] != "p1" { t.Fatalf("synth names: p0=%v p1=%v", p0["name"], p1["name"]) }
                    if p0["type"] != "int" || p1["type"] != "int" { t.Fatalf("synth types: p0=%v p1=%v", p0["type"], p1["type"]) }
                    found = true
                }
            }
        }
    }
    if !found { t.Fatalf("External call with synthesized params not found in IR") }
}

