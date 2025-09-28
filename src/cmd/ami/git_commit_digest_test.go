package main

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "testing"
)

// TestComputeCommitDigest_LocalRepo_Sha1Format verifies that for a SHA-1 formatted local git repo,
// computeCommitDigest returns the SHA-256 of the canonical raw commit object content
// ("commit <len>\x00" + body), matching our independent calculation.
func TestComputeCommitDigest_LocalRepo_Sha1Format(t *testing.T) {
    if os.Getenv("AMI_E2E_ENABLE_GIT") != "1" { t.Skip("set AMI_E2E_ENABLE_GIT=1 to enable git-backed tests") }
    if _, err := exec.LookPath("git"); err != nil { t.Skip("git not found") }
    {
        ctx, cancel := context.WithTimeout(context.Background(), 5_000_000_000)
        defer cancel()
        if err := exec.CommandContext(ctx, "git", "--version").Run(); err != nil || ctx.Err() != nil {
            t.Skip("git --version failed; skipping")
        }
    }

    dir := filepath.Join("build", "test", "git_commit_digest", "repo")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    run := func(args ...string) {
        ctx, cancel := context.WithTimeout(context.Background(), 30_000_000_000)
        defer cancel()
        cmd := exec.CommandContext(ctx, "git", args...)
        cmd.Dir = dir
        cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
        if out, err := cmd.CombinedOutput(); err != nil { t.Fatalf("git %v: %v\n%s", args, err, out) }
    }
    run("init")
    if err := os.WriteFile(filepath.Join(dir, "x.txt"), []byte("hello"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    run("add", ".")
    run("-c", "user.email=test@example.com", "-c", "user.name=test", "commit", "-m", "init")
    run("tag", "v0.1.0")

    // compute via helper
    got, err := computeCommitDigest(dir, "v0.1.0")
    if err != nil { t.Fatalf("computeCommitDigest: %v", err) }
    if len(got) != 64 { t.Fatalf("expected 64-hex digest, got %q", got) }

    // independently compute canonical SHA-256 from raw commit body
    // Resolve commit id
    out, err := exec.Command("git", "-C", dir, "rev-parse", "v0.1.0^{commit}").CombinedOutput()
    if err != nil { t.Fatalf("rev-parse: %v: %s", err, string(out)) }
    // body
    body, err := exec.Command("git", "-C", dir, "cat-file", "-p", string(out)).CombinedOutput()
    if err != nil { t.Fatalf("cat-file: %v: %s", err, string(body)) }
    header := []byte("commit " + strconv.Itoa(len(body)) + "\x00")
    h := sha256.New()
    _, _ = h.Write(header)
    _, _ = h.Write(body)
    want := hex.EncodeToString(h.Sum(nil))
    if got != want { t.Fatalf("digest mismatch: got=%s want=%s", got, want) }
}

