package driver

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Tail self-call should be marked with LOOP/GOTO in IR debug to indicate TCE.
func TestLower_TailCallElimination_Self(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    // Minimal tail recursion: return F()
    code := "package app\nfunc F(){ return F() }\n"
    fs.AddFile("tce.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    arts, _ := Compile(ws, pkgs, Options{Debug: true})
    if len(arts.IR) == 0 { t.Fatalf("no IR emitted") }
    b, err := os.ReadFile(arts.IR[0])
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    fns := obj["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions") }
    fn := fns[0].(map[string]any)
    blks := fn["blocks"].([]any)
    blk := blks[0].(map[string]any)
    instrs := blk["instrs"].([]any)
    hasLoop := false
    hasGoto := false
    for _, in := range instrs {
        m := in.(map[string]any)
        if m["op"] == "LOOP" { hasLoop = true }
        if m["op"] == "GOTO" && m["label"] == "entry" { hasGoto = true }
    }
    if !hasLoop || !hasGoto { t.Fatalf("expected LOOP and GOTO entry markers; got: %+v", instrs) }
}

