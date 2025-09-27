package main

import (
    "bytes"
    "encoding/json"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// Validate that ami.sum sha256 uses commit digest, not directory hash.
func TestModGet_Git_ShaIsCommitDigest(t *testing.T) {
    repo := filepath.Join("build", "test", "mod_get_commit", "repo")
    ws := filepath.Join("build", "test", "mod_get_commit", "ws")
    cache := filepath.Join("build", "test", "mod_get_commit", "cache")
    _ = os.RemoveAll(filepath.Join("build", "test", "mod_get_commit"))
    if err := os.MkdirAll(repo, 0o755); err != nil { t.Fatalf("mkdir repo: %v", err) }
    run := func(name string, args ...string) {
        cmd := exec.Command(name, args...)
        cmd.Dir = repo
        cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
        if out, err := cmd.CombinedOutput(); err != nil { t.Fatalf("%s %v: %v\n%s", name, args, err, out) }
    }
    run("git", "init")
    if err := os.WriteFile(filepath.Join(repo, "x.txt"), []byte("commit-digest"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    run("git", "add", ".")
    run("git", "-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-m", "c1")
    run("git", "tag", "v1.2.3")

    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir ws: %v", err) }
    if err := os.WriteFile(filepath.Join(ws, "ami.workspace"), []byte("version: 1.0.0\npackages: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    old := os.Getenv("AMI_PACKAGE_CACHE")
    defer os.Setenv("AMI_PACKAGE_CACHE", old)
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)

    absRepo, _ := filepath.Abs(repo)
    var buf bytes.Buffer
    if err := runModGet(&buf, ws, "file+git://"+absRepo+"#v1.2.3", true); err != nil { t.Fatalf("runModGet: %v", err) }
    // Read ami.sum and compare digest
    b, err := os.ReadFile(filepath.Join(ws, "ami.sum"))
    if err != nil { t.Fatalf("read sum: %v", err) }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil { t.Fatalf("json sum: %v", err) }
    pkgs := m["packages"].(map[string]any)
    repoObj := pkgs["repo"].(map[string]any)
    got := repoObj["sha256"].(string)

    exp, err := computeCommitDigest(repo, "v1.2.3")
    if err != nil { t.Fatalf("computeCommitDigest: %v", err) }
    if got != exp {
        t.Fatalf("sha mismatch: got=%s want=%s", got, exp)
    }
}

