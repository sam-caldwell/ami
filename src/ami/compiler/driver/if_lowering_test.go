package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestDriver_IfLowering_CondBrBlocks(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ if (1 == 2) { return } else { return } }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns, _ := obj["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions in IR") }
    // Check entry has a COND_BR and blocks include then/else/join labels
    fn := fns[0].(map[string]any)
    blks, _ := fn["blocks"].([]any)
    if len(blks) < 3 { t.Fatalf("expected multiple blocks, got %d", len(blks)) }
    entry := blks[0].(map[string]any)
    instrs, _ := entry["instrs"].([]any)
    sawCond := false
    for _, in := range instrs {
        m := in.(map[string]any)
        if m["op"] == "COND_BR" { sawCond = true; break }
    }
    if !sawCond { t.Fatalf("missing COND_BR in entry: %s", string(b)) }
}

