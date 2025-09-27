package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestPipelineVisualize_FilterByFile_JSON(t *testing.T) {
    dir := t.TempDir()
    // workspace
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "a.ami"), []byte("package app\npipeline P() { ingress; egress }\n"), 0o644); err != nil { t.Fatalf("write a: %v", err) }
    if err := os.WriteFile(filepath.Join(srcDir, "b.ami"), []byte("package app\npipeline Q() { ingress; egress }\n"), 0o644); err != nil { t.Fatalf("write b: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    sel := filepath.Join("src", "a.ami")
    c.SetArgs([]string{"pipeline", "visualize", "--json", "--file", sel})
    // execute in temp dir
    ol, _ := os.Getwd()
    _ = os.Chdir(dir)
    defer os.Chdir(ol)
    if err := c.Execute(); err != nil { t.Fatalf("execute: %v", err) }
    s := out.String()
    if !bytes.Contains(out.Bytes(), []byte("\"name\":\"P\"")) || bytes.Contains(out.Bytes(), []byte("\"name\":\"Q\"")) {
        t.Fatalf("filter not applied; out=%s", s)
    }
}

