package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "testing"
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
        cmd := exec.Command(c[0], c[1:]...)
        out, err := cmd.CombinedOutput()
        if err != nil { t.Fatalf("%v: %s", err, string(out)) }
    }
    dig, err := computeCommitDigest(dir, "HEAD")
    if err != nil { t.Fatalf("computeCommitDigest: %v", err) }
    if len(dig) != 64 || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(dig) {
        t.Fatalf("unexpected digest: %q", dig)
    }
}

