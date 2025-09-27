package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestObjectStub_SchemaContainsSymbolsAndRelocs(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nfunc F(){}\nfunc G(){}\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    of := filepath.Join("build", "obj", "app", "u.o")
    b, err := os.ReadFile(of)
    if err != nil { t.Fatalf("read: %v", err) }
    var obj map[string]any
    if e := json.Unmarshal(b, &obj); e != nil { t.Fatalf("json: %v", e) }
    if _, ok := obj["symbols"].([]any); !ok { t.Fatalf("symbols missing or wrong type: %T", obj["symbols"]) }
    if _, ok := obj["relocs"].([]any); !ok { t.Fatalf("relocs missing or wrong type: %T", obj["relocs"]) }
}

