package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testLint_StrictFlag_PromotesWarnings(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_lint", "strict_flag")
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.ami"), []byte("UNKNOWN_IDENT\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Packages[0].Package.Root = "./src"
	ws.Toolchain.Linter.Options = []string{} // workspace not strict
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	// CLI strict flag = true; expect error in JSON mode
	var buf bytes.Buffer
	if err := runLint(&buf, dir, true, false, true); err == nil {
		// runLint returns error only when errors>0; strict should promote warn to error
		// decode and verify summary errors>0 if not returned error
		dec := json.NewDecoder(&buf)
		var errors float64
		for dec.More() {
			var m map[string]any
			if derr := dec.Decode(&m); derr != nil {
				t.Fatalf("json: %v", derr)
			}
			if m["code"] == "SUMMARY" {
				if e, ok := m["data"].(map[string]any); ok {
					errors = e["errors"].(float64)
				}
			}
		}
		if errors == 0 {
			t.Fatalf("expected errors>0 under strict; out=%s", buf.String())
		}
	}
}

func TestLint_Help_ShowsStageBFlags(t *testing.T) {
	cmd := newLintCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()
	s := buf.String()
	if !bytes.Contains([]byte(s), []byte("stage-b")) {
		t.Fatalf("expected help to mention stage-b; got: %s", s)
	}
}
