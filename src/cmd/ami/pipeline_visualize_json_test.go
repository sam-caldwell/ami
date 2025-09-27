package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestPipelineVisualize_JSON_EmitsGraphV1(t *testing.T) {
    dir := t.TempDir()
    // workspace with main root
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // simple pipeline file
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package app\n\npipeline P() {\n  ingress().transform().egress()\n}\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"pipeline", "visualize", "--json"})
    c.SetIn(bytes.NewReader(nil))
    // execute in temp dir
    ol, _ := os.Getwd()
    _ = os.Chdir(dir)
    defer os.Chdir(ol)
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    b := out.Bytes()
    if !bytes.Contains(b, []byte(`"schema":"graph.v1"`)) {
        t.Fatalf("missing schema: %s", string(b))
    }
    if !bytes.Contains(b, []byte(`"name":"P"`)) {
        t.Fatalf("missing pipeline name: %s", string(b))
    }
    if !bytes.Contains(b, []byte(`"type":"summary"`)) {
        t.Fatalf("missing summary record: %s", string(b))
    }
}

