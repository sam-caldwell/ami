package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Mutual recursion tail calls should produce LOOP/DISPATCH/GOTO markers.
func TestLower_TailCallElimination_Mutual(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc A(){ return B() }\nfunc B(){ return A() }\n"
    fs.AddFile("mut.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, _ := Compile(ws, pkgs, Options{Debug: true})
    if len(arts.IR) == 0 { t.Fatalf("no IR emitted") }
    b, err := os.ReadFile(arts.IR[0])
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    fns := obj["functions"].([]any)
    if len(fns) < 2 { t.Fatalf("expected two functions") }
    // Find function A IR and check for DISPATCH B
    var a map[string]any
    for _, it := range fns { m := it.(map[string]any); if m["name"] == "A" { a = m; break } }
    if a == nil { t.Fatalf("function A not found") }
    blk := a["blocks"].([]any)[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    hasLoop := false; hasDispatch := false; hasGoto := false
    for _, in := range instrs {
        m := in.(map[string]any)
        if m["op"] == "LOOP" { hasLoop = true }
        if m["op"] == "DISPATCH" && m["label"] == "B" { hasDispatch = true }
        if m["op"] == "GOTO" && m["label"] == "entry" { hasGoto = true }
    }
    if !hasLoop || !hasDispatch || !hasGoto { t.Fatalf("expected LOOP, DISPATCH B, and GOTO entry in A; got: %+v", instrs) }
}

