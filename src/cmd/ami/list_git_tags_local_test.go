package main

import (
    "context"
    "os/exec"
    "path/filepath"
    "testing"
    stdtime "time"
)

func Test_listGitTags_localRepo(t *testing.T) {
    if _, err := exec.LookPath("git"); err != nil {
        t.Skip("git not available; skipping local repo tag test")
    }
    dir := t.TempDir()
    abs := filepath.Clean(dir)
    // Initialize a minimal git repo and create two tags
    cmds := [][]string{
        {"git", "init", "-q"},
        {"git", "config", "user.email", "test@example.com"},
        {"git", "config", "user.name", "Test"},
        {"git", "-c", "commit.gpgSign=false", "commit", "--allow-empty", "-m", "init"},
        {"git", "-c", "tag.gpgSign=false", "tag", "v1.2.3"},
        {"git", "-c", "tag.gpgSign=false", "tag", "v2.0.0"},
    }
    for _, c := range cmds {
        ctx, cancel := context.WithTimeout(context.Background(), 5*stdtime.Second)
        defer cancel()
        cmd := exec.CommandContext(ctx, c[0], c[1:]...)
        cmd.Dir = abs
        if out, err := cmd.CombinedOutput(); err != nil {
            t.Fatalf("%v: %v\n%s", c, err, string(out))
        }
    }
    tags, err := listGitTags(abs)
    if err != nil { t.Fatalf("listGitTags: %v", err) }
    seen := map[string]bool{}
    for _, tname := range tags { seen[tname] = true }
    if !seen["v1.2.3"] || !seen["v2.0.0"] {
        t.Fatalf("expected v1.2.3 and v2.0.0 in %v", tags)
    }
}
