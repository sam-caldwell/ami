package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func testLint_TODO_FIXME_Warns(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "todo_fixme")
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	wsYAML := []byte("---\nversion: 1.0.0\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: [\"darwin/arm64\"]\n  linker:\n    options: [\"Optimize: 0\"]\n  linter:\n    options: []\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
	if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), wsYAML, 0o644); err != nil {
		t.Fatal(err)
	}
	src := []byte("// TODO: implement\n/* FIXME: broken */\n")
	if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runLint(&buf, dir, true, false, false); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("\"code\":\"W_TODO\"")) {
		t.Fatalf("expected W_TODO in output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("\"code\":\"W_FIXME\"")) {
		t.Fatalf("expected W_FIXME in output: %s", out)
	}
}

func testLint_TODO_PragmaDisable(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "todo_disable")
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	wsYAML := []byte("---\nversion: 1.0.0\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: [\"darwin/arm64\"]\n  linker:\n    options: [\"Optimize: 0\"]\n  linter:\n    options: []\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
	if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), wsYAML, 0o644); err != nil {
		t.Fatal(err)
	}
	src := []byte("#pragma lint:disable W_TODO\n// TODO: will be suppressed\n/* FIXME: still warn */\n")
	if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := runLint(&buf, dir, true, false, false); err != nil {
		t.Fatalf("unexpected error: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if bytes.Contains([]byte(out), []byte("\"code\":\"W_TODO\"")) {
		t.Fatalf("did not expect W_TODO in output: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("\"code\":\"W_FIXME\"")) {
		t.Fatalf("expected W_FIXME in output: %s", out)
	}
}
