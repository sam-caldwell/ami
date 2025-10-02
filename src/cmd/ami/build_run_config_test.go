package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testRunBuild_UsesWorkspaceConfig_JSON(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_build", "config_json")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Compiler.Target = "./out"
	ws.Toolchain.Compiler.Env = []string{"linux/amd64", "darwin/arm64"}
	// No remote imports; focus on config extraction
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	var buf bytes.Buffer
	if err := runBuild(&buf, dir, true, false); err != nil {
		t.Fatalf("runBuild: %v", err)
	}
	var m map[string]any
	if e := json.Unmarshal(buf.Bytes(), &m); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	data := m["data"].(map[string]any)
	if data == nil {
		t.Fatalf("expected data present")
	}
	if data["targetDir"] == "" {
		t.Fatalf("expected targetDir populated")
	}
	targets, ok := data["targets"].([]any)
	if !ok || len(targets) != 2 {
		t.Fatalf("expected 2 targets; got %v", data["targets"])
	}
}

func testRunBuild_DefaultEnvWhenEmpty_JSON(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_build", "default_env")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Compiler.Env = nil
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}

	var buf bytes.Buffer
	if err := runBuild(&buf, dir, true, false); err != nil {
		t.Fatalf("runBuild: %v", err)
	}
	var m map[string]any
	if e := json.Unmarshal(buf.Bytes(), &m); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	data := m["data"].(map[string]any)
	if data == nil {
		t.Fatalf("expected data present")
	}
	targets := data["targets"].([]any)
	if len(targets) != 1 {
		t.Fatalf("expected default single target; got %v", targets)
	}
}
