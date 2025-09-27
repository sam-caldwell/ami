package main

import (
    "bytes"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// Ensure mod get selects the highest non-prerelease tag when none is provided.
func TestModGet_SelectsHighestTag_WhenOmitted(t *testing.T) {
    base := filepath.Join("build", "test", "mod_get", "select_latest")
    repo := filepath.Join(base, "repo")
    ws := filepath.Join(base, "ws")
    cache := filepath.Join(base, "cache")
    _ = os.RemoveAll(base)
    if err := os.MkdirAll(repo, 0o755); err != nil { t.Fatalf("mkdir repo: %v", err) }
    // Init repo and create content
    run := func(name string, args ...string) {
        cmd := exec.Command(name, args...)
        cmd.Dir = repo
        cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
        if out, err := cmd.CombinedOutput(); err != nil { t.Fatalf("%s %v: %v\n%s", name, args, err, out) }
    }
    run("git", "init")
    if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    run("git", "add", ".")
    run("git", "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-m", "init")
    run("git", "tag", "v1.2.0")
    run("git", "tag", "v1.3.0-rc.1")
    run("git", "tag", "v1.3.0")

    // Minimal workspace
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir ws: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }

    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    // Invoke mod get using file+git without a tag
    absRepo, _ := filepath.Abs(repo)
    src := "file+git://" + absRepo // no #tag
    var buf bytes.Buffer
    if err := runModGet(&buf, ws, src, true); err != nil { t.Fatalf("runModGet: %v", err) }
    var res modGetResult
    if err := json.Unmarshal(buf.Bytes(), &res); err != nil { t.Fatalf("json: %v; out=%s", err, buf.String()) }
    if res.Version != "v1.3.0" {
        t.Fatalf("expected version v1.3.0, got %s (out=%s)", res.Version, buf.String())
    }
    if _, err := os.Stat(filepath.Join(cache, "repo", "v1.3.0", "a.txt")); err != nil { t.Fatalf("expected cached repo@v1.3.0: %v", err) }
}

