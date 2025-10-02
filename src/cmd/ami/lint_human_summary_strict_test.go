package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testLint_Human_Summary_And_StrictPromotion(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "human_summary")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "package x\n// TODO: fix\nfunc F(){}\n"
	if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Packages[0].Package.Root = "./src"
	ws.Toolchain.Linter.Options = []string{}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Human, non-strict: expect warnings summary line and exit nil
	var out bytes.Buffer
	if err := runLint(&out, dir, false, false, false); err != nil {
		t.Fatalf("runLint: %v", err)
	}
	s := out.String()
	if !bytes.Contains([]byte(s), []byte("lint:")) {
		t.Fatalf("expected human lint output, got: %s", s)
	}

	// Human, strict: warnings promoted to errors and non-nil error
	out.Reset()
	if err := runLint(&out, dir, false, false, true); err == nil {
		t.Fatalf("expected error in strict mode with warnings")
	}
}

func testLint_Human_MaxWarn_Boundary(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "human_maxwarn")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "package x\n// TODO: one\n// TODO: two\n"
	if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Packages[0].Package.Root = "./src"
	ws.Toolchain.Linter.Options = []string{}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Set maxWarn=1; with two TODOs expect failure in JSON mode
	setLintOptions(LintOptions{MaxWarn: 1})
	defer setLintOptions(LintOptions{MaxWarn: -1})
	var out bytes.Buffer
	if err := runLint(&out, dir, true, false, false); err == nil {
		t.Fatalf("expected error when warnings exceed maxWarn; out=%s", out.String())
	}
}
