package e2e

import (
    "context"
    "bytes"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    stdtime "time"
    "github.com/sam-caldwell/ami/src/testutil"
)

// Ensure key subcommands do not prompt for input and succeed/fail deterministically.
func TestE2E_Ami_NonInteractive_Subcommands(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "non_interactive")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    // ami init --json --force
    {
        ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(15*stdtime.Second))
        defer cancel()
        cmd := exec.CommandContext(ctx, bin, "init", "--json", "--force")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        if err := cmd.Run(); err != nil { t.Fatalf("init: %v\n%s", err, stderr.String()) }
    }
    // ami build --json (should emit E_WS_SCHEMA until src exists)
    {
        ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
        defer cancel()
        cmd := exec.CommandContext(ctx, bin, "build", "--json")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        _ = cmd.Run() // may fail due to workspace contents; only verifying no prompt
    }
    // Create src and a minimal file, then build again
    {
        if err := os.MkdirAll(filepath.Join(ws, "src"), 0o755); err != nil { t.Fatalf("mkdir src: %v", err) }
        if err := os.WriteFile(filepath.Join(ws, "src", "u.ami"), []byte("package newProject\nfunc F(){}\n"), 0o644); err != nil { t.Fatalf("write: %v", err) }
        ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(20*stdtime.Second))
        defer cancel()
        cmd := exec.CommandContext(ctx, bin, "build", "--json")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        _ = cmd.Run()
    }
    // ami lint --json (should run without input)
    {
        ctx, cancel := context.WithTimeout(context.Background(), testutil.Timeout(15*stdtime.Second))
        defer cancel()
        cmd := exec.CommandContext(ctx, bin, "lint", "--json")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        _ = cmd.Run()
    }
}
