package main

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
    stdtime "time"
)

// Ensure listGitTags handles the remote path codepath by pointing git to a local repo via a relative path.
func Test_listGitTags_remotePathAgainstLocalRepo(t *testing.T) {
    if _, err := exec.LookPath("git"); err != nil {
        t.Skip("git not available; skipping ls-remote tags test")
    }
    // Create a local repo with both lightweight and annotated tags
    dir := t.TempDir()
    // On Windows, relative paths can be tricky with drive letters; normalize later
    cmds := [][]string{
        {"git", "init", "-q"},
        {"git", "config", "user.email", "test@example.com"},
        {"git", "config", "user.name", "Test"},
        {"bash", "-lc", "echo hello > f.txt"},
        {"git", "add", "f.txt"},
        {"git", "-c", "commit.gpgSign=false", "commit", "-m", "init"},
        // lightweight tag
        {"git", "-c", "tag.gpgSign=false", "tag", "v0.1.0"},
        // annotated tag (ensure we strip ^{} entries)
        {"git", "-c", "tag.gpgSign=false", "tag", "-a", "-m", "release", "v0.2.0"},
    }
    for _, c := range cmds {
        ctx, cancel := context.WithTimeout(context.Background(), 5*stdtime.Second)
        defer cancel()
        cmd := exec.CommandContext(ctx, c[0], c[1:]...)
        cmd.Dir = dir
        if out, err := cmd.CombinedOutput(); err != nil {
            t.Fatalf("%v: %v\n%s", c, err, string(out))
        }
    }
    // Use a relative path to trigger the ls-remote code path
    cwd, _ := os.Getwd()
    rel, err := filepath.Rel(cwd, dir)
    if err != nil { t.Fatalf("rel: %v", err) }
    // On Windows, ensure rel is not absolute
    if runtime.GOOS == "windows" && filepath.IsAbs(rel) {
        // Fallback: use "." then join to make a relative-ish path in the same drive
        rel = filepath.Base(dir)
    }
    tags, err := listGitTags(rel)
    if err != nil { t.Fatalf("listGitTags: %v", err) }
    seen := map[string]bool{}
    for _, s := range tags { seen[s] = true }
    if !seen["v0.1.0"] || !seen["v0.2.0"] {
        t.Fatalf("expected tags v0.1.0 and v0.2.0 in %v", tags)
    }
}

