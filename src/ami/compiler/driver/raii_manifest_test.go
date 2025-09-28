package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestManifest_Includes_RAII_Debug_Path(t *testing.T) {
    fs := &source.FileSet{}
    code := "package app\nfunc F(){ var x string; release(x); defer release(x) }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(workspace.Workspace{}, pkgs, Options{Debug: true})
    mf := filepath.Join("build", "debug", "manifest.json")
    b, err := os.ReadFile(mf)
    if err != nil { t.Fatalf("manifest read: %v", err) }
    var obj struct{ Packages []struct{ Name string; Units []struct{ Unit, RAII string } } }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    found := false
    for _, p := range obj.Packages {
        if p.Name != "app" { continue }
        for _, u := range p.Units {
            if u.Unit == "u" && u.RAII != "" {
                if _, err := os.Stat(u.RAII); err == nil { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected RAII path in manifest") }
}

