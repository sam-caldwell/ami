package driver

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

func TestIR_Capabilities_FromNETUsage_IngressEgress(t *testing.T) {
    ws := workspace.Workspace{}
    fs := &source.FileSet{}
    code := "package app\npipeline P(){ ingress; net.Connect(); egress }\n"
    fs.AddFile("u.ami", code)
    pkgs := []Package{{Name: "app", Files: fs}}
    _, _ = Compile(ws, pkgs, Options{Debug: true})
    path := filepath.Join("build", "debug", "ir", "app", "u.ir.json")
    b, err := os.ReadFile(path)
    if err != nil { t.Fatalf("read ir: %v", err) }
    var obj map[string]any
    if err := json.Unmarshal(b, &obj); err != nil { t.Fatalf("json: %v", err) }
    caps, _ := obj["capabilities"].([]any)
    if len(caps) == 0 { t.Fatalf("capabilities missing: %#v", obj["capabilities"]) }
    found := false
    for _, c := range caps { if c.(string) == "net" { found = true } }
    if !found { t.Fatalf("expected net cap in IR: %v", caps) }
}

