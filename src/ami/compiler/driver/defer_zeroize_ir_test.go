package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify that defer release(a) lowers to a DEFER containing a call to ami_rt_zeroize_owned in IR debug JSON.
func TestLower_DeferRelease_EmitsZeroizeOwned_InIR(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var a Owned; defer release(a) }\n"
    fs.AddFile("d.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    irPath := filepath.Join("build", "debug", "ir", "app", "d.ir.json")
    b, err := os.ReadFile(irPath)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil { t.Fatalf("json: %v", err) }
    fns := m["functions"].([]any)
    if len(fns) == 0 { t.Fatalf("no functions in IR") }
    fn0 := fns[0].(map[string]any)
    blks := fn0["blocks"].([]any)
    if len(blks) == 0 { t.Fatalf("no blocks in IR") }
    instrs := blks[0].(map[string]any)["instrs"].([]any)
    found := false
    for _, ii := range instrs {
        obj := ii.(map[string]any)
        if obj["op"] == "DEFER" {
            ex := obj["expr"].(map[string]any)
            if ex["op"] == "call" && ex["callee"] == "ami_rt_zeroize_owned" {
                found = true
                break
            }
        }
    }
    if !found { t.Fatalf("DEFER zeroize-owned not found in IR: %s", string(b)) }
}

