package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testModGet_LocalDeclared_CopiesAndUpdatesSum(t *testing.T) {
	// Workspace setup
	wsdir := filepath.Join("build", "test", "mod_get", "ws")
	if err := os.MkdirAll(filepath.Join(wsdir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create workspace with additional package
	sumPath := filepath.Join(wsdir, "ami.workspace")
	content := []byte("---\nversion: 1.0.0\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: [\"darwin/arm64\"]\n    options: [\"verbose\"]\n  linker:\n    options: [\"Optimize: 0\"]\n  linter:\n    options: [\"strict\"]\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n  - util:\n      name: util\n      version: 1.2.3\n      root: ./util\n      import: []\n")
	if err := os.WriteFile(sumPath, content, 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	// Seed util package content
	if err := os.MkdirAll(filepath.Join(wsdir, "util"), 0o755); err != nil {
		t.Fatalf("mkdir util: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "util", "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Point cache to temp dir
	cache := filepath.Join("build", "test", "mod_get", "cache")
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)

	// Run
	var buf bytes.Buffer
	if err := runModGet(&buf, wsdir, "./util", true); err != nil {
		t.Fatalf("runModGet: %v", err)
	}
	var res modGetResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v; out=%s", err, buf.String())
	}
	if res.Name != "util" || res.Version != "1.2.3" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(cache, "util", "1.2.3", "file.txt")); err != nil {
		t.Fatalf("expected cached file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(wsdir, "ami.sum")); err != nil {
		t.Fatalf("expected ami.sum created: %v", err)
	}
}

func testModGet_LocalDeclared_HumanOutput(t *testing.T) {
	wsdir := filepath.Join("build", "test", "mod_get", "ws_human")
	if err := os.MkdirAll(filepath.Join(wsdir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte("---\nversion: 1.0.0\npackages:\n  - util:\n      name: util\n      version: 2.0.0\n      root: ./util\n      import: []\n")
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), content, 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(wsdir, "util"), 0o755); err != nil {
		t.Fatalf("mkdir util: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "util", "f.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cache := filepath.Join("build", "test", "mod_get", "cache_human")
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)
	var buf bytes.Buffer
	if err := runModGet(&buf, wsdir, "./util", false); err != nil {
		t.Fatalf("runModGet: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("fetched util@2.0.0")) {
		t.Fatalf("expected human output; got: %s", buf.String())
	}
}

func testModGet_Errors_WhenOutsideWorkspace(t *testing.T) {
	wsdir := filepath.Join("build", "test", "mod_get", "err_outside")
	if err := os.MkdirAll(filepath.Join(wsdir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	var buf bytes.Buffer
	if err := runModGet(&buf, wsdir, "../other", true); err == nil {
		t.Fatalf("expected error for outside path")
	}
}

func testModGet_Errors_WhenNotDeclared(t *testing.T) {
	wsdir := filepath.Join("build", "test", "mod_get", "err_undeclared")
	if err := os.MkdirAll(filepath.Join(wsdir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte("---\nversion: 1.0.0\npackages:\n  - key: main\n    package:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n")
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), content, 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	// Make a folder not declared
	if err := os.MkdirAll(filepath.Join(wsdir, "lib"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	if err := runModGet(&buf, wsdir, "./lib", true); err == nil {
		t.Fatalf("expected error for undeclared path")
	}
}
