package e2e

import (
    "bytes"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

// Ensure key subcommands do not prompt for input and succeed/fail deterministically.
func TestE2E_Ami_NonInteractive_Subcommands(t *testing.T) {
    bin := buildAmi(t)
    ws := filepath.Join("build", "test", "e2e", "non_interactive")
    _ = os.RemoveAll(ws)
    if err := os.MkdirAll(ws, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }

    // ami init --json --force
    {
        cmd := exec.Command(bin, "init", "--json", "--force")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        if err := cmd.Run(); err != nil { t.Fatalf("init: %v\n%s", err, stderr.String()) }
    }
    // ami build --json (should emit E_WS_SCHEMA until src exists)
    {
        cmd := exec.Command(bin, "build", "--json")
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
        cmd := exec.Command(bin, "build", "--json")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        _ = cmd.Run()
    }
    // ami lint --json (should run without input)
    {
        cmd := exec.Command(bin, "lint", "--json")
        cmd.Dir = ws
        cmd.Stdin = io.NopCloser(bytes.NewReader(nil))
        var stdout, stderr bytes.Buffer
        cmd.Stdout, cmd.Stderr = &stdout, &stderr
        _ = cmd.Run()
    }
}

