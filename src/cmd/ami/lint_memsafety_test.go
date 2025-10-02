package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

// Validates that memory safety diagnostics are emitted by Stage B.
func testLint_MemorySafety_EmitsDiagnostics_JSON(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "memsafe")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Introduce banned '&' and unsupported unary '*' usage
	content := "a := &b\n*c + d\n* x = y\n" // third line is allowed pattern and should not error
	if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Packages[0].Package.Root = "./src"
	// Non-strict mode; memory safety emits errors by default
	ws.Toolchain.Linter.Options = []string{}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Enable Stage B memory safety
	setRuleToggles(RuleToggles{StageB: true, MemorySafety: true})
	defer setRuleToggles(RuleToggles{})

	var buf bytes.Buffer
	err := runLint(&buf, dir, true, false, false)
	if err == nil {
		// With errors, runLint returns non-nil in JSON mode only when errors>0; we check codes anyway
	}
	dec := json.NewDecoder(&buf)
	var sawPtr, sawMut bool
	for dec.More() {
		var m map[string]any
		if derr := dec.Decode(&m); derr != nil {
			t.Fatalf("json: %v", derr)
		}
		if m["code"] == "E_PTR_UNSUPPORTED_SYNTAX" {
			sawPtr = true
		}
		if m["code"] == "E_MUT_BLOCK_UNSUPPORTED" {
			sawMut = true
		}
	}
	if !sawPtr || !sawMut {
		t.Fatalf("expected memory safety diagnostics; out=%s", buf.String())
	}
}

// Validates that pragmas can disable memory safety diagnostics per file.
func testLint_MemorySafety_PragmaDisable(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "memsafe_pragma")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "#pragma lint:disable E_PTR_UNSUPPORTED_SYNTAX\na := &b\n"
	if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Packages[0].Package.Root = "./src"
	ws.Toolchain.Linter.Options = []string{}
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	setRuleToggles(RuleToggles{StageB: true, MemorySafety: true})
	defer setRuleToggles(RuleToggles{})
	var buf bytes.Buffer
	if err := runLint(&buf, dir, true, false, false); err != nil {
		// ignore
	}
	dec := json.NewDecoder(&buf)
	for dec.More() {
		var m map[string]any
		if derr := dec.Decode(&m); derr != nil {
			t.Fatalf("json: %v", derr)
		}
		if m["code"] == "E_PTR_UNSUPPORTED_SYNTAX" {
			t.Fatalf("pragma failed to suppress ptr diag: %s", buf.String())
		}
	}
}
