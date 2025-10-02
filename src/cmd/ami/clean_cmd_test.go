package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sam-caldwell/ami/src/ami/exit"
	"github.com/sam-caldwell/ami/src/ami/logging"
	"github.com/sam-caldwell/ami/src/ami/workspace"
)

func testRunClean_FreshRepo_DefaultBuild(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_clean", "fresh")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	target := filepath.Join(dir, "build")
	_ = os.RemoveAll(target)

	var buf bytes.Buffer
	if err := runClean(&buf, dir, false); err != nil {
		t.Fatalf("runClean: %v", err)
	}
	st, err := os.Stat(target)
	if err != nil || !st.IsDir() {
		t.Fatalf("expected build dir created; err=%v st=%v", err, st)
	}
}

func testRunClean_WithExistingFiles_RemovesAndRecreates(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_clean", "existing")
	target := filepath.Join(dir, "build")
	nested := filepath.Join(target, "tmp.txt")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(nested, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var buf bytes.Buffer
	if err := runClean(&buf, dir, false); err != nil {
		t.Fatalf("runClean: %v", err)
	}
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatalf("expected nested file to be removed; err=%v", err)
	}
	if st, err := os.Stat(target); err != nil || !st.IsDir() {
		t.Fatalf("expected build dir exists; err=%v st=%v", err, st)
	}
}

func testRunClean_JSONOutputAndWorkspaceTarget(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_clean", "json")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Write a workspace with custom target
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Compiler.Target = "./out"
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save workspace: %v", err)
	}
	var buf bytes.Buffer
	if err := runClean(&buf, dir, true); err != nil {
		t.Fatalf("runClean: %v", err)
	}
	var res cleanResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v; out=%s", err, buf.String())
	}
	if !res.Removed || !res.Created {
		t.Fatalf("expected removed and created true; res=%+v", res)
	}
	if _, err := os.Stat(filepath.Join(dir, "out")); err != nil {
		t.Fatalf("expected custom target created: %v", err)
	}
}

func testRunClean_AbsoluteTarget_Errors(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_clean", "abs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	var abs string
	if runtime.GOOS == "windows" {
		// Simulate an absolute-like path on Windows tests (avoid requiring a specific drive)
		abs = filepath.Join(string(filepath.Separator), "ami_test_abs")
	} else {
		abs = "/tmp/ami_test_abs"
	}
	ws.Toolchain.Compiler.Target = abs
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	var buf bytes.Buffer
	err := runClean(&buf, dir, true)
	if err == nil {
		t.Fatalf("expected error for absolute target")
	}
	if exit.UnwrapCode(err) != exit.User {
		t.Fatalf("expected User exit code; got %v", exit.UnwrapCode(err))
	}
	// Verify JSON emitted even on error path
	var res cleanResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json decode: %v; out=%s", e, buf.String())
	}
	if res.Path == "" {
		t.Fatalf("expected res.Path populated for absolute target error")
	}
}

func testRunClean_DefaultTarget_MessageRecorded(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_clean", "default_msg")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// No workspace file present
	var buf bytes.Buffer
	if err := runClean(&buf, dir, true); err != nil {
		t.Fatalf("runClean: %v", err)
	}
	var res cleanResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json decode: %v; out=%s", e, buf.String())
	}
	if len(res.Messages) == 0 {
		t.Fatalf("expected messages to include default target note")
	}
}

func testRunClean_MkdirFailsWhenParentIsFile(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_clean", "mkdir_fail_parent_file")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Create a parent file so that target parent is not a directory
	if err := os.WriteFile(filepath.Join(dir, "parent"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write parent: %v", err)
	}
	ws := workspace.DefaultWorkspace()
	ws.Toolchain.Compiler.Target = "parent/child"
	if err := ws.Save(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("save: %v", err)
	}
	var buf bytes.Buffer
	err := runClean(&buf, dir, true)
	if err == nil {
		t.Fatalf("expected error when parent is a file")
	}
	var res cleanResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json decode: %v; out=%s", e, buf.String())
	}
	if res.Created {
		t.Fatalf("expected created=false on mkdir failure")
	}
}

func testRunClean_LogsToDebugWhenVerbose(t *testing.T) {
	base := filepath.Join("build", "test", "ami_clean", "debug_logs")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lg, _ := logging.New(logging.Options{JSON: true, Verbose: true, DebugDir: base})
	setRootLogger(lg)
	defer closeRootLogger()

	dir := filepath.Join(base, "ws")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	var buf bytes.Buffer
	if err := runClean(&buf, dir, true); err != nil {
		t.Fatalf("runClean: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(base, "activity.log"))
	if err != nil {
		t.Fatalf("read activity.log: %v", err)
	}
	if !bytes.Contains(data, []byte("clean.start")) || !bytes.Contains(data, []byte("clean.created")) {
		t.Fatalf("expected clean debug lines; got: %s", string(data))
	}
}
