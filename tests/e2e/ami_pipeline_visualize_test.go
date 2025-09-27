package e2e

import (
    "bytes"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func TestAmiPipelineVisualize_JSON_SchemaPresent(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "pipeline_visualize", "json")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    wsy := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), wsy, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    src := []byte("package app\n\npipeline P() { ingress; transform; egress }\n")
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), src, 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(bin, "pipeline", "visualize", "--json")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    if !strings.Contains(stdout.String(), "\"schema\":\"graph.v1\"") {
        t.Fatalf("missing schema in output: %s", stdout.String())
    }
}

func TestAmiPipelineVisualize_ASCII_Line(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "pipeline_visualize", "ascii")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    wsy := []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), wsy, 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    src := []byte("package app\n\npipeline P() { ingress; transform; egress }\n")
    if err := os.WriteFile(filepath.Join(ws, "src", "main.ami"), src, 0o644); err != nil { t.Fatalf("write src: %v", err) }
    cmd := exec.Command(bin, "pipeline", "visualize")
    cmd.Dir = ws
    cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
    var stdout, stderr bytes.Buffer
    cmd.Stdout, cmd.Stderr = &stdout, &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("run: %v\nstderr=%s\nstdout=%s", err, stderr.String(), stdout.String())
    }
    s := stdout.String()
    if !strings.Contains(s, "package: app  pipeline: P\n") || !strings.Contains(s, "[ingress] --> (transform) --> [egress]\n") {
        t.Fatalf("unexpected ascii output: %q", s)
    }
}
