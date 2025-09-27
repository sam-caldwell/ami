package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestPipelineVisualize_ASCII_RendersHeaderAndLine(t *testing.T) {
    dir := t.TempDir()
    // workspace with main root
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    // simple pipeline file
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    content := "package app\n\npipeline P() {\n  ingress; transform; egress\n}\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"pipeline", "visualize"})
    c.SetIn(bytes.NewReader(nil))
    // execute in temp dir
    ol, _ := os.Getwd()
    _ = os.Chdir(dir)
    defer os.Chdir(ol)
    if err := c.Execute(); err != nil {
        t.Fatalf("execute: %v", err)
    }
    got := out.String()
    if !bytes.Contains(out.Bytes(), []byte("package: app  pipeline: P\n")) {
        t.Fatalf("missing header: %q", got)
    }
    if !bytes.Contains(out.Bytes(), []byte("[ingress] --> (transform) --> [egress]\n")) {
        t.Fatalf("missing ascii line: %q", got)
    }
}
