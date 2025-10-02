package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func testModGet_GitFileScheme_TagsAndCopies(t *testing.T) {
	// Guard: require explicit opt-in and a working git
	if os.Getenv("AMI_E2E_ENABLE_GIT") != "1" {
		t.Skip("git tests disabled; set AMI_E2E_ENABLE_GIT=1 to enable")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5_000_000_000)
		defer cancel()
		cmd := exec.CommandContext(ctx, "git", "--version")
		if err := cmd.Run(); err != nil || ctx.Err() != nil {
			t.Skip("git --version failed or timed out; skipping")
		}
	}
	// Set up a local git repo
	repo := filepath.Join("build", "test", "mod_get_git", "repo")
	_ = os.RemoveAll(repo)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	run := func(name string, args ...string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30_000_000_000)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}
	run("git", "init")
	if err := os.WriteFile(filepath.Join(repo, "x.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run("git", "add", ".")
	run("git", "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-m", "init")
	run("git", "tag", "v1.2.3")

	// workspace receiving ami.sum
	wsdir := filepath.Join("build", "test", "mod_get_git", "ws")
	if err := os.MkdirAll(wsdir, 0o755); err != nil {
		t.Fatalf("mkdir ws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}

	cache := filepath.Join("build", "test", "mod_get_git", "cache")
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)

	// Use file+git scheme with absolute path
	absRepo, _ := filepath.Abs(repo)
	src := "file+git://" + absRepo + "#v1.2.3"
	var buf bytes.Buffer
	if err := runModGet(&buf, wsdir, src, true); err != nil {
		t.Fatalf("runModGet: %v", err)
	}
	var res modGetResult
	if err := json.Unmarshal(buf.Bytes(), &res); err != nil {
		t.Fatalf("json: %v; out=%s", err, buf.String())
	}
	if res.Name != "repo" || res.Version != "v1.2.3" {
		t.Fatalf("unexpected: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(cache, "repo", "v1.2.3", "x.txt")); err != nil {
		t.Fatalf("cached file missing: %v", err)
	}
	// Ensure VCS metadata was not copied
	if _, err := os.Stat(filepath.Join(cache, "repo", "v1.2.3", ".git")); err == nil {
		t.Fatalf(".git directory should not be copied to cache")
	}
	if _, err := os.Stat(filepath.Join(wsdir, "ami.sum")); err != nil {
		t.Fatalf("ami.sum missing: %v", err)
	}
}

func testModGet_GitSSH_MissingTag_Errors(t *testing.T) {
	// No actual network call should occur because missing tag is validated early.
	wsdir := filepath.Join("build", "test", "mod_get_git", "ssh_missing_tag")
	if err := os.MkdirAll(wsdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	var buf bytes.Buffer
	src := "git+ssh://git@github.com/org/repo.git" // no #tag
	if err := runModGet(&buf, wsdir, src, true); err == nil {
		t.Fatalf("expected error for missing tag")
	}
}

func testModGet_FileGit_RequiresAbsolutePath(t *testing.T) {
	wsdir := filepath.Join("build", "test", "mod_get_git", "file_rel")
	if err := os.MkdirAll(wsdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	var buf bytes.Buffer
	src := "file+git://relative/path#v1.0.0" // relative path is invalid
	if err := runModGet(&buf, wsdir, src, true); err == nil {
		t.Fatalf("expected error for relative file+git path")
	}
}
