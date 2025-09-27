package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestIR_Capabilities_Trust_FromPragmas(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma capabilities list=io,network\n#pragma trust level=untrusted\nfunc F(){}\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    caps, ok := obj["capabilities"].([]any)
    if !ok || len(caps) == 0 { t.Fatalf("capabilities missing: %#v", obj["capabilities"]) }
    if obj["trustLevel"].(string) != "untrusted" { t.Fatalf("trustLevel: %v", obj["trustLevel"]) }
}

