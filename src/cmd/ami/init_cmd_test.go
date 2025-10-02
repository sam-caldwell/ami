package main

import (
	"bytes"
	"encoding/json"
	"github.com/sam-caldwell/ami/src/ami/logging"
	"os"
	"path/filepath"
	"testing"
)

func testRunInit_CreatesWorkspaceAndDirs(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "create")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	if err := runInit(&buf, dir, true, true); err != nil { // use --force to bypass git requirement
		t.Fatalf("runInit: %v", err)
	}
	// Verify JSON output parses
	var res initResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v; out=%s", err, buf.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("workspace not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "build")); err != nil {
		t.Fatalf("target dir not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src")); err != nil {
		t.Fatalf("package dir not created: %v", err)
	}
}

func testRunInit_ForceAddsMissingFields(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "force_missing")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Write a minimal broken file
	if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var buf bytes.Buffer
	if err := runInit(&buf, dir, true, true); err != nil {
		t.Fatalf("runInit(force): %v", err)
	}
	var res initResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if !res.Updated {
		t.Fatalf("expected updated=true when filling missing fields")
	}
	// Validate that directories got created and .gitignore updated
	if _, err := os.Stat(filepath.Join(dir, "build")); err != nil {
		t.Fatalf("target dir not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "src")); err != nil {
		t.Fatalf("package dir not created: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if !bytes.Contains(b, []byte("./build")) {
		t.Fatalf(".gitignore missing ./build entry")
	}
}

func testRunInit_ErrorsWhenNotGitRepoWithoutForce(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "not_git")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	err := runInit(&buf, dir, false, false)
	if err == nil {
		t.Fatalf("expected error when not in git repo without --force")
	}
}

func testRunInit_ForceNoChangesIsIdempotent(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "idempotent")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	// First run creates the workspace
	if err := runInit(&buf, dir, true, false); err != nil {
		t.Fatalf("first run: %v", err)
	}
	// Second run with --force should detect no missing fields and not rewrite materially
	buf.Reset()
	if err := runInit(&buf, dir, true, false); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("workspace file present")) {
		t.Fatalf("expected idempotent message, got: %s", buf.String())
	}
}

func testRunInit_ErrorPathEmitsJSONWhenRequested(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "json_error")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	err := runInit(&buf, dir, false, true)
	if err == nil {
		t.Fatalf("expected error when not git repo")
	}
	var res initResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("expected JSON on error path: %v", e)
	}
}

func testRunInit_DoesNotDuplicateGitignoreLine(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "gitignore")
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil { // make it a git repo
		t.Fatalf("mkdir .git: %v", err)
	}
	gi := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gi, []byte("./build\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	var buf bytes.Buffer
	if err := runInit(&buf, dir, false, false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	b, _ := os.ReadFile(gi)
	count := bytes.Count(b, []byte("./build\n"))
	if count != 1 {
		t.Fatalf("expected single ./build line, got %d", count)
	}
}

func testRunInit_ForceInitializesGit(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "git_init")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	if err := runInit(&buf, dir, true, true); err != nil {
		t.Fatalf("runInit(force): %v", err)
	}
	var res initResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	// At least one of these states should hold after --force
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Fatalf("expected .git to exist after --force; err=%v", err)
	}
	if res.GitStatus != "initialized" && res.GitStatus != "present" {
		t.Fatalf("unexpected git status: %s", res.GitStatus)
	}
}

func testRunInit_LogsToDebugWhenVerbose(t *testing.T) {
	// Set a root logger with JSON to simplify matching, writing under a test dir.
	base := filepath.Join("build", "test", "ami_init", "debug_logs")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	lg, _ := logging.New(logging.Options{JSON: true, Verbose: true, DebugDir: base})
	setRootLogger(lg)
	defer closeRootLogger()

	dir := filepath.Join(base, "workspace")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}
	var buf bytes.Buffer
	if err := runInit(&buf, dir, true, true); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(base, "activity.log"))
	if err != nil {
		t.Fatalf("read activity.log: %v", err)
	}
	if !bytes.Contains(data, []byte("init.start")) || !bytes.Contains(data, []byte("init.complete")) {
		t.Fatalf("expected debug lines in activity.log; got: %s", string(data))
	}
}

func testRunInit_ForceGitNotFound_YieldsRequiredStatus(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "git_not_found")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Ensure not a git repo, and force path where git is not found
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	_ = os.Setenv("PATH", "/nonexistent")

	var buf bytes.Buffer
	if err := runInit(&buf, dir, true, true); err != nil {
		t.Fatalf("runInit(force,json): %v", err)
	}
	var res initResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if res.GitStatus != "required" {
		t.Fatalf("expected gitStatus=required when git missing, got %s", res.GitStatus)
	}
}

func testRunInit_HumanCreatedMessage(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "human_created")
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil { // treat as git repo
		t.Fatalf("mkdir .git: %v", err)
	}
	_ = os.Remove(filepath.Join(dir, "ami.workspace"))
	var buf bytes.Buffer
	if err := runInit(&buf, dir, false, false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("created workspace file")) {
		t.Fatalf("expected created message, got: %s", out)
	}
}

func testRunInit_HumanUpdatedMessage(t *testing.T) {
	dir := filepath.Join("build", "test", "ami_init", "human_updated")
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil { // treat as git repo
		t.Fatalf("mkdir .git: %v", err)
	}
	// Minimal file missing fields
	if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var buf bytes.Buffer
	if err := runInit(&buf, dir, true, false); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("updated workspace file (missing fields)")) {
		t.Fatalf("expected updated message, got: %s", out)
	}
}

func testRootCmd_WiresInitSubcommand(t *testing.T) {
	t.Skip("pending: stabilize root/working-dir integration under cobra for CI")
	dir := filepath.Join("build", "test", "ami_init", "root_integration")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"init", "--json", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var res initResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("workspace not created by root/init: %v", err)
	}
}

func testNewInitCmd_ExecuteDirect(t *testing.T) {
	t.Skip("pending: stabilize direct cobra Execute() integration test")
	dir := filepath.Join("build", "test", "ami_init", "direct_init")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	c := newInitCmd()
	var buf bytes.Buffer
	c.SetOut(&buf)
	c.SetArgs([]string{"--json", "--force"})
	if err := c.Execute(); err != nil {
		t.Fatalf("execute init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ami.workspace")); err != nil {
		t.Fatalf("workspace not created: %v", err)
	}
}

func testNewRootCmd_ConstructsAndHasFlags(t *testing.T) {
	c := newRootCmd()
	if c == nil {
		t.Fatalf("newRootCmd returned nil")
	}
	if f := c.PersistentFlags().Lookup("json"); f == nil {
		t.Fatalf("missing --json flag")
	}
	if f := c.PersistentFlags().Lookup("verbose"); f == nil {
		t.Fatalf("missing --verbose flag")
	}
	if f := c.PersistentFlags().Lookup("color"); f == nil {
		t.Fatalf("missing --color flag")
	}
}
