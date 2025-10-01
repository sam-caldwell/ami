package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify ir.symbols.index.json includes ami_rt_signal_register when signal.Register is used.
func TestIRSymbols_Index_Includes_SignalRegister(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nimport signal\nfunc H(){}\nfunc F(){ signal.Register(signal.SIGTERM, H) }\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true, EmitLLVMOnly: true})
    idx := filepath.Join("build", "debug", "ir", "app", "ir.symbols.index.json")
    b, err := os.ReadFile(idx)
    if err != nil { t.Fatalf("read symbols index: %v", err) }
    var obj struct{
        Schema string `json:"schema"`
        Package string `json:"package"`
        Units []struct{
            Unit string `json:"unit"`
            Externs []string `json:"externs"`
        } `json:"units"`
    }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if obj.Package != "app" { t.Fatalf("package mismatch: %s", obj.Package) }
    found := false
    for _, u := range obj.Units {
        for _, e := range u.Externs {
            if e == "ami_rt_signal_register" { found = true }
        }
    }
    if !found { t.Fatalf("expected ami_rt_signal_register extern in index: %s", string(b)) }
}

