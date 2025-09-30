package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestIR_Includes_EventMeta_Block(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    fs.AddFile("u.ami", "package app\nfunc F(){}\n")
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    em, ok := obj["eventmeta"].(map[string]any)
    if !ok { t.Fatalf("eventmeta missing: %v", obj) }
    if em["schema"] != "eventmeta.v1" { t.Fatalf("schema: %v", em["schema"]) }
    fields, ok := em["fields"].([]any)
    if !ok || len(fields) == 0 { t.Fatalf("fields: %v", em["fields"]) }
}

