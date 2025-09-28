package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify manifest carries IR, LLVM, and RAII paths for a compiled unit.
func TestManifest_Contains_IR_LLVM_RAII_Trio_ForUnit(t *testing.T) {
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var x string; release(x); defer release(x) }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(workspace.Workspace{}, pkgs, Options{Debug: true})

    mf := filepath.Join("build", "debug", "manifest.json")
    b, err := os.ReadFile(mf)
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var obj struct{
        Packages []struct{
            Name  string
            Units []struct{ Unit, IR, LLVM, RAII string }
        }
    }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }

    found := false
    for _, p := range obj.Packages {
        if p.Name != "app" { continue }
        for _, u := range p.Units {
            if u.Unit == "u" {
                if u.IR == "" || u.LLVM == "" || u.RAII == "" {
                    t.Fatalf("missing paths: IR=%q LLVM=%q RAII=%q", u.IR, u.LLVM, u.RAII)
                }
                if _, err := os.Stat(u.IR); err != nil { t.Fatalf("ir missing: %v", err) }
                if _, err := os.Stat(u.LLVM); err != nil { t.Fatalf("llvm missing: %v", err) }
                if _, err := os.Stat(u.RAII); err != nil { t.Fatalf("raii missing: %v", err) }
                found = true
                break
            }
        }
    }
    if !found { t.Fatalf("expected IR/LLVM/RAII trio for unit 'u'") }
}

