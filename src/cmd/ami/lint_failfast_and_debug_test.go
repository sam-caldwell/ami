package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testLint_Human_FailFast_WarnsReturnError(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "human_failfast")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "package x\n// TODO: one\n"
	if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Packages[0].Package.Root = "./src"
	ws.Toolchain.Linter.Options = []string{}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	setLintOptions(LintOptions{FailFast: true})
	defer setLintOptions(LintOptions{})
	var out bytes.Buffer
	if err := runLint(&out, dir, false, false, false); err == nil {
		t.Fatalf("expected error in human mode with failfast on warning; out=%s", out.String())
	}
}

func testLint_DebugFile_NotCreated_WhenNotVerbose(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "debug_not_verbose")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Linter.Options = []string{}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	var out bytes.Buffer
	if err := runLint(&out, dir, true, false, false); err != nil { /* ok */
	}
	if _, err := os.Stat(filepath.Join(dir, "build", "debug", "lint.ndjson")); err == nil {
		t.Fatalf("debug file should not exist when not verbose")
	}
}
