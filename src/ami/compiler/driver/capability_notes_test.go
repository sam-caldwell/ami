package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestContracts_CapabilityNotes_GenericAndSpecificIO(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\n#pragma capabilities list=io\npipeline P(){ ingress; io.Read(\"f\"); egress }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.contracts.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read contracts: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    caps, _ := obj["capabilities"].([]any)
    hasIO := false
    hasRead := false
    for _, c := range caps {
        if c.(string) == "io" { hasIO = true }
        if c.(string) == "io.read" { hasRead = true }
    }
    if !(hasIO && hasRead) { t.Fatalf("expected both 'io' and 'io.read' in caps: %v", caps) }
    notes, _ := obj["capabilityNotes"].([]any)
    if len(notes) == 0 { t.Fatalf("expected capabilityNotes present: %v", obj) }
}

