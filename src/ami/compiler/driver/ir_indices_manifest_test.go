package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Verify build manifest references IR indices (types/symbols) when debug is enabled.
func TestManifest_Includes_IR_Indices(t *testing.T) {
    ws := workspace.Workspace{}
    var fs source.FileSet
    fs.AddFile("u.ami", "package app\nfunc F(){}\n")
    pkgs := []Package{{Name: "app", Files: &fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})

    mf := filepath.Join("build", "debug", "manifest.json")
    b, err := os.ReadFile(mf)
    if err != nil { t.Fatalf("read manifest: %v", err) }
    var obj struct{ Packages []struct{ Name string; IRTypesIndex, IRSymbolsIndex string } }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    if len(obj.Packages) == 0 { t.Fatalf("packages empty in manifest") }
    found := false
    for _, p := range obj.Packages {
        if p.Name != "app" { continue }
        if p.IRTypesIndex == "" || p.IRSymbolsIndex == "" {
            t.Fatalf("missing IR indices in manifest: types=%q symbols=%q", p.IRTypesIndex, p.IRSymbolsIndex)
        }
        // ensure referenced files exist
        if _, err := os.Stat(p.IRTypesIndex); err != nil { t.Fatalf("types index missing: %v", err) }
        if _, err := os.Stat(p.IRSymbolsIndex); err != nil { t.Fatalf("symbols index missing: %v", err) }
        found = true
    }
    if !found { t.Fatalf("package app not found in manifest") }
}

