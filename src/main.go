package main

import (
    "os"
    "os/exec"
)

// Temporary adapter: delegate to the full CLI under ./src/cmd/ami via `go run`.
// This keeps the build target `src/main.go` working while the CLI lives under cmd/ami.
func main() {
    args := os.Args[1:]
    cmd := exec.Command("go", append([]string{"run", "./src/cmd/ami"}, args...)...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin
    if err := cmd.Run(); err != nil {
        // best-effort exit code propagation
        if ee, ok := err.(*exec.ExitError); ok {
            os.Exit(ee.ExitCode())
        }
        os.Exit(1)
    }
}
