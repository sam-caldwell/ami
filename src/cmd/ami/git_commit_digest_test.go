package main

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

func TestComputeCommitDigest_SimpleRepo(t *testing.T) {
    dir := filepath.Join("build", "test", "git_digest")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Initialize repo and make one commit
    cmds := [][]string{
        {"git", "-C", dir, "init"},
        {"git", "-C", dir, "config", "user.email", "test@example.com"},
        {"git", "-C", dir, "config", "user.name", "Test"},
        {"bash", "-lc", "echo hello > " + filepath.Join(dir, "f.txt")},
        {"git", "-C", dir, "add", "f.txt"},
        {"git", "-C", dir, "commit", "-m", "init"},
    }
    for _, c := range cmds {
        ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(5*stdtime.Second))
        defer cancel()
        cmd := exec.CommandContext(ctx, c[0], c[1:]...)
        out, err := cmd.CombinedOutput()
        if err != nil { t.Fatalf("%v: %s", err, string(out)) }
    }
    dig, err := computeCommitDigest(dir, "HEAD")
    if err != nil { t.Fatalf("computeCommitDigest: %v", err) }
    if len(dig) != 64 || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(dig) {
        t.Fatalf("unexpected digest: %q", dig)
    }
}

func TestComputeCommitDigest_InvalidRef(t *testing.T) {
    dir := filepath.Join("build", "test", "git_digest_badref")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Initialize empty repo
    ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(5*stdtime.Second))
    defer cancel()
    if out, err := exec.CommandContext(ctx, "git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
        t.Fatalf("git init: %v: %s", err, string(out))
    }
    if _, err := computeCommitDigest(dir, "does-not-exist"); err == nil {
        t.Fatal("expected error for invalid ref")
    }
}
