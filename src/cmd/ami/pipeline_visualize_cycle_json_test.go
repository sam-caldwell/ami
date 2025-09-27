package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestPipelineVisualize_JSON_Cycle_EmitsDiagAndErrors(t *testing.T) {
    dir := t.TempDir()
    // workspace
    ws := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), ws, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    srcDir := filepath.Join(dir, "src")
    if err := os.MkdirAll(srcDir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // pipeline with cycle: A -> B; B -> A
    content := "package app\n\npipeline P() {\n  A; B; A -> B; B -> A;\n}\n"
    if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil { t.Fatalf("write: %v", err) }

    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs([]string{"pipeline", "visualize", "--json"})
    c.SetIn(bytes.NewReader(nil))
    ol, _ := os.Getwd()
    _ = os.Chdir(dir)
    defer os.Chdir(ol)
    if err := c.Execute(); err == nil {
        t.Fatalf("expected error due to cycle")
    }
    if !bytes.Contains(out.Bytes(), []byte(`"code":"E_GRAPH_CYCLE"`)) {
        t.Fatalf("expected E_GRAPH_CYCLE diag; out=%s", out.String())
    }
}

