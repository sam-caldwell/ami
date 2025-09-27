package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestLLVM_DebugWrite_AndManifestReference(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\nfunc F() { }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    ll := filepath.Join("build", "debug", "llvm", "app", "u.ll")
    b, err := os.ReadFile(ll)
    if err != nil { t.Fatalf("read llvm: %v", err) }
    if len(b) == 0 { t.Fatalf("expected non-empty llvm dump") }
    // manifest contains path
    mf := filepath.Join("build", "debug", "manifest.json")
    mb, err := os.ReadFile(mf)
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var obj struct{ Packages []struct{ Name string; Units []struct{ Unit, LLVM string } } }
    if err := json.Unmarshal(mb, &obj); err != nil { t.Fatalf("json: %v", err) }
    found := false
    for _, p := range obj.Packages {
        if p.Name != "app" { continue }
        for _, u := range p.Units {
            if u.Unit == "u" && u.LLVM == ll { found = true; break }
        }
    }
    if !found { t.Fatalf("manifest missing llvm path: %s", ll) }
}

