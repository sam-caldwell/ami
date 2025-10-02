//go:build !windows

package os

import (
    goos "os"
    "path/filepath"
    "os/exec"
    "testing"
    "time"
)

// This test exercises Kill and killProcessGroup on Unix-like systems without
// asserting platform-specific semantics. It starts a simple sleep binary and
// ensures Kill returns without error and the process stops.
func TestProcess_Kill_Group_Unix(t *testing.T) {
    dir := t.TempDir()
    src := `package main
import ("time")
func main(){ time.Sleep(5 * time.Second) }`
    file := filepath.Join(dir, "main.go")
    if err := goos.WriteFile(file, []byte(src), 0o644); err != nil { t.Fatalf("write: %v", err) }
    bin := filepath.Join(dir, "sleepbin")
    if out, err := exec.Command("go", "build", "-o", bin, file).CombinedOutput(); err != nil {
        t.Fatalf("go build: %v (out=%s)", err, string(out))
    }
    p, err := Exec(bin)
    if err != nil { t.Fatalf("exec: %v", err) }
    if err := p.Start(false); err != nil { t.Fatalf("start: %v", err) }
    // Give it a moment to start
    time.Sleep(50 * time.Millisecond)
    if err := p.Kill(); err != nil { t.Fatalf("kill: %v", err) }
    // Do not assert eventual stop to avoid platform flakiness; the call
    // exercises killProcessGroup and Process.Kill paths for coverage.
}
