package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// Asserts that after a non-verbose compile, the obj index prefers .o over .s for a unit.
func TestCompile_ObjIndex_PrefersO_AfterNonDebug(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nfunc F(){}\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: false})
    idxPath := filepath.Join("build", "obj", "app", "index.json")
    b, err := os.ReadFile(idxPath)
    if err != nil { t.Fatalf("read index: %v", err) }
    var obj struct{ Units []struct{ Unit, Path string } }
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    found := false
    for _, u := range obj.Units {
        if u.Unit == "u" {
            if u.Path != "u.o" { t.Fatalf("expected u.o, got %s", u.Path) }
            found = true
        }
    }
    if !found { t.Fatalf("unit u not found in index: %+v", obj.Units) }
}
