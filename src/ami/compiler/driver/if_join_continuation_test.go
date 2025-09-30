package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that statements after an if/else are lowered into a join block and preserved.
func TestDriver_IfLowering_JoinContinuation(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var x int; if (1 == 2) { x = 1 } else { x = 2 }; return x }\n"
    fs.AddFile("u2.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    path := filepath.Join("build", "debug", "ir", "app", "u2.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    fns, _ := obj["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions in IR") }
    fn := fns[0].(map[string]any)
    blks, _ := fn["blocks"].([]any)
    // Expect at least 4 blocks: entry, then, else, join
    if len(blks) < 4 { t.Fatalf("expected >=4 blocks, got %d", len(blks)) }
    // Find join block and ensure it contains a RETURN
    hasJoinReturn := false
    for _, bb := range blks {
        m := bb.(map[string]any)
        if m["name"] == "join0" || m["name"] == "join1" { // tolerate id variation
            ins, _ := m["instrs"].([]any)
            for _, in := range ins {
                if im, ok := in.(map[string]any); ok && im["op"] == "RETURN" { hasJoinReturn = true }
            }
        }
    }
    if !hasJoinReturn { t.Fatalf("expected RETURN in join block; got: %s", string(b)) }
}

