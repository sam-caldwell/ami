package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

func TestLint_LangNotGo_Warns(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_lint", "lang_not_go")
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatal(err) }
    wsYAML := []byte("---\nversion: 1.0.0\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: [\"darwin/arm64\"]\n  linker:\n    options: [\"Optimize: 0\"]\n  linter:\n    options: []\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), wsYAML, 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), []byte(""), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "misc.go"), []byte("package x\n"), 0o644); err != nil { t.Fatal(err) }
    var buf bytes.Buffer
    if err := runLint(&buf, dir, true, false, false); err != nil { t.Fatalf("unexpected error: %v\n%s", err, buf.String()) }
    out := buf.String()
    if !bytes.Contains([]byte(out), []byte("\"code\":\"W_LANG_NOT_GO\"")) { t.Fatalf("expected W_LANG_NOT_GO in output: %s", out) }
}

