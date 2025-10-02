package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func testModSum_MissingFile_JSON(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "missing")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var buf bytes.Buffer
	err := runModSum(&buf, dir, true)
	if err == nil {
		t.Fatalf("expected error for missing ami.sum")
	}
	var res modSumResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	if res.Ok {
		t.Fatalf("expected ok=false")
	}
}

func testModSum_InvalidJSON(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "invalid")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var buf bytes.Buffer
	if err := runModSum(&buf, dir, true); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func testModSum_BadSchema(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "bad_schema")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte(`{"schema":"wrong","packages":[]}`)
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var buf bytes.Buffer
	if err := runModSum(&buf, dir, true); err == nil {
		t.Fatalf("expected error for wrong schema")
	}
}

func testModSum_Happy_Minimal(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "happy")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte(`{"schema":"ami.sum/v1","packages":{}}`)
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), content, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	var buf bytes.Buffer
	if err := runModSum(&buf, dir, true); err != nil {
		t.Fatalf("runModSum: %v", err)
	}
	var res modSumResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	if !res.Ok || res.PackagesSeen != 0 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func testModSum_IntegrityMismatch(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "mismatch")
	cache := filepath.Join(dir, "cache")
	if err := os.MkdirAll(filepath.Join(cache, "alpha", "1.0.0"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cache, "alpha", "1.0.0", "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Point env at test cache
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)
	// Wrong hash
	sum := []byte(`{"schema":"ami.sum/v1","packages":[{"name":"alpha","version":"1.0.0","sha256":"000000"}]}`)
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}
	var buf bytes.Buffer
	err := runModSum(&buf, dir, true)
	if err == nil {
		t.Fatalf("expected integrity error")
	}
}

func testModSum_IntegrityMatch(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "match")
	cache := filepath.Join(dir, "cache")
	p := filepath.Join(cache, "beta", "2.0.0")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(p, "a.txt"), []byte("A"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(p, "b.txt"), []byte("B"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)
	// Compute expected hash using same logic as implementation
	got, err := hashDir(p)
	if err != nil {
		t.Fatalf("hashDir: %v", err)
	}
	sum := []byte(fmt.Sprintf(`{"schema":"ami.sum/v1","packages":[{"name":"beta","version":"2.0.0","sha256":"%s"}]}`, got))
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}
	var buf bytes.Buffer
	if err := runModSum(&buf, dir, true); err != nil {
		t.Fatalf("runModSum: %v", err)
	}
	var res modSumResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	if !res.Ok || res.PackagesSeen != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func testModSum_FetchesMissing_FromFileGit_AndUpdatesSum(t *testing.T) {
	if os.Getenv("AMI_E2E_ENABLE_GIT") != "1" {
		t.Skip("git tests disabled; set AMI_E2E_ENABLE_GIT=1 to enable")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5_000_000_000)
		defer cancel()
		if err := exec.CommandContext(ctx, "git", "--version").Run(); err != nil || ctx.Err() != nil {
			t.Skip("git --version failed; skipping")
		}
	}
	base := filepath.Join("build", "test", "mod_sum", "fetch_git")
	repo := filepath.Join(base, "repo")
	wsdir := filepath.Join(base, "ws")
	cache := filepath.Join(base, "cache")
	// set up repo
	_ = os.RemoveAll(base)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	run := func(dir string, name string, args ...string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30_000_000_000)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}
	run(repo, "git", "init")
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run(repo, "git", "add", ".")
	run(repo, "git", "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-m", "init")
	run(repo, "git", "tag", "v1.0.0")

	// workspace and sum
	if err := os.MkdirAll(wsdir, 0o755); err != nil {
		t.Fatalf("mkdir ws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wsdir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}
	absRepo, _ := filepath.Abs(repo)
	// ami.sum with wrong sha256 and file+git source; package name derived as "repo"
	sum := []byte(fmt.Sprintf(`{"schema":"ami.sum/v1","packages":{"repo":{"version":"v1.0.0","sha256":"0000","source":"file+git://%s"}}}`, absRepo))
	if err := os.WriteFile(filepath.Join(wsdir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}

	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)

	var buf bytes.Buffer
	if err := runModSum(&buf, wsdir, true); err != nil {
		t.Fatalf("runModSum: %v\n%s", err, buf.String())
	}
	// after run, cache should contain repo/v1.0.0 and sum should be updated with correct sha
	if _, err := os.Stat(filepath.Join(cache, "repo", "v1.0.0", "a.txt")); err != nil {
		t.Fatalf("cache missing: %v", err)
	}
	// read sum and verify sha updated
	b, err := os.ReadFile(filepath.Join(wsdir, "ami.sum"))
	if err != nil {
		t.Fatalf("read sum: %v", err)
	}
	var m map[string]any
	if json.Unmarshal(b, &m) != nil {
		t.Fatalf("sum not json: %s", string(b))
	}
	pkgs := m["packages"].(map[string]any)
	repoObj := pkgs["repo"].(map[string]any)
	if repoObj["sha256"] == "0000" {
		t.Fatalf("sha256 not updated")
	}
}

func testModSum_WorkspaceCrossCheck_MissingInSum(t *testing.T) {
	dir := filepath.Join("build", "test", "mod_sum", "ws_miss")
	cache := filepath.Join(dir, "cache")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatalf("mkdir cache: %v", err)
	}

	// Create workspace with two packages
	wsYAML := []byte("---\nversion: 1.0.0\ntoolchain:\n  compiler:\n    concurrency: NUM_CPU\n    target: ./build\n    env: [\"darwin/arm64\"]\n    options: [\"verbose\"]\n  linker:\n    options: [\"Optimize: 0\"]\n  linter:\n    options: [\"strict\"]\npackages:\n  - alpha:\n      name: alpha\n      version: 1.0.0\n      root: ./alpha\n      import: []\n  - beta:\n      name: beta\n      version: 2.0.0\n      root: ./beta\n      import: []\n")
	if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), wsYAML, 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}

	// Populate cache for both alpha and beta
	pAlpha := filepath.Join(cache, "alpha", "1.0.0")
	pBeta := filepath.Join(cache, "beta", "2.0.0")
	if err := os.MkdirAll(pAlpha, 0o755); err != nil {
		t.Fatalf("mkdir alpha: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pAlpha, "a.txt"), []byte("A"), 0o644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.MkdirAll(pBeta, 0o755); err != nil {
		t.Fatalf("mkdir beta: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pBeta, "b.txt"), []byte("B"), 0o644); err != nil {
		t.Fatalf("write beta: %v", err)
	}

	// Compute alpha hash and create ami.sum with only alpha present
	hAlpha, err := hashDir(pAlpha)
	if err != nil {
		t.Fatalf("hash alpha: %v", err)
	}
	sum := []byte(fmt.Sprintf(`{"schema":"ami.sum/v1","packages":{"alpha":{"version":"1.0.0","sha256":"%s"}}}`, hAlpha))
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}

	// Point AMI_PACKAGE_CACHE at our test cache
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)

	var buf bytes.Buffer
	err = runModSum(&buf, dir, true)
	if err == nil {
		t.Fatalf("expected non-nil error due to missing beta in sum")
	}
	var res modSumResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	// alpha verified, beta should be in missing due to absent in ami.sum
	wantMiss := "beta@2.0.0"
	found := false
	for _, m := range res.Missing {
		if m == wantMiss {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected missing to include %s; got %+v", wantMiss, res)
	}
}

func testModSum_GitFetchMissingInCache_UpdatesSum(t *testing.T) {
	if os.Getenv("AMI_E2E_ENABLE_GIT") != "1" {
		t.Skip("git tests disabled; set AMI_E2E_ENABLE_GIT=1 to enable")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5_000_000_000)
		defer cancel()
		if err := exec.CommandContext(ctx, "git", "--version").Run(); err != nil || ctx.Err() != nil {
			t.Skip("git --version failed; skipping")
		}
	}
	// Set up a local git repo
	repo := filepath.Join("build", "test", "mod_sum_git", "repo")
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

	// Prepare workspace and sum referencing git source
	dir := filepath.Join("build", "test", "mod_sum_git", "ws")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir ws: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil {
		t.Fatalf("write ws: %v", err)
	}

	// sum: object form with source and empty sha
	absRepo, _ := filepath.Abs(repo)
	sum := []byte(fmt.Sprintf(`{"schema":"ami.sum/v1","packages":{"repo":{"version":"v1.2.3","sha256":"","source":"file+git://%s"}}}`, absRepo))
	if err := os.WriteFile(filepath.Join(dir, "ami.sum"), sum, 0o644); err != nil {
		t.Fatalf("write sum: %v", err)
	}

	cache := filepath.Join("build", "test", "mod_sum_git", "cache")
	old := os.Getenv("AMI_PACKAGE_CACHE")
	defer os.Setenv("AMI_PACKAGE_CACHE", old)
	_ = os.Setenv("AMI_PACKAGE_CACHE", cache)

	var buf bytes.Buffer
	if err := runModSum(&buf, dir, true); err != nil {
		t.Fatalf("runModSum: %v", err)
	}
	var res modSumResult
	if e := json.Unmarshal(buf.Bytes(), &res); e != nil {
		t.Fatalf("json: %v; out=%s", e, buf.String())
	}
	if !res.Ok {
		t.Fatalf("expected ok after fetch; res=%+v", res)
	}
	if _, err := os.Stat(filepath.Join(cache, "repo", "v1.2.3", "x.txt")); err != nil {
		t.Fatalf("cached file missing: %v", err)
	}
	// Verify ami.sum updated with sha256
	b, err := os.ReadFile(filepath.Join(dir, "ami.sum"))
	if err != nil {
		t.Fatalf("read sum: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("json sum: %v", err)
	}
	pkgs := m["packages"].(map[string]any)
	repoObj := pkgs["repo"].(map[string]any)
	if strOrEmpty(repoObj["sha256"]) == "" {
		t.Fatalf("expected sha256 to be populated after fetch")
	}
}
