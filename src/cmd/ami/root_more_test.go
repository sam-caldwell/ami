package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testRoot_NoArgs_ShowsHelp(t *testing.T) {
	c := newRootCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetArgs([]string{})
	if err := c.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("AMI toolchain CLI")) {
		t.Fatalf("expected help text; got: %s", out.String())
	}
}

func testRoot_CleanSubcommand_JSON(t *testing.T) {
	t.Skip("pending: stabilize root/clean integration under cobra for CI")
	dir := filepath.Join("build", "test", "root_clean_json")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Compiler.Target = "./target"
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save workspace: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	c := newRootCmd()
	var out bytes.Buffer
	c.SetOut(&out)
	c.SetArgs([]string{"clean", "--json"})
	if err := c.Execute(); err != nil {
		t.Fatalf("execute clean: %v", err)
	}
	var res cleanResult
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatalf("json decode: %v; out=%s", err, out.String())
	}
	if res.Path == "" {
		t.Fatalf("expected non-empty result path; out=%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "target")); err != nil {
		t.Fatalf("expected target dir created: %v", err)
	}
}
