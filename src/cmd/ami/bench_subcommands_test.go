package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// benchSandbox creates an isolated workspace directory and configures environment
// to avoid polluting the user environment (e.g., AMI_PACKAGE_CACHE under temp dir).
func benchSandbox(b *testing.B) (string, func()) {
    b.Helper()
    dir := b.TempDir()
    cache := filepath.Join(dir, "cache")
    _ = os.MkdirAll(cache, 0o755)
    oldCache := os.Getenv("AMI_PACKAGE_CACHE")
    _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    // disable any potential interactive prompts from child tools (e.g., git)
    oldGitPrompt := os.Getenv("GIT_TERMINAL_PROMPT")
    _ = os.Setenv("GIT_TERMINAL_PROMPT", "0")
    cleanup := func() {
        _ = os.Setenv("AMI_PACKAGE_CACHE", oldCache)
        _ = os.Setenv("GIT_TERMINAL_PROMPT", oldGitPrompt)
    }
    return dir, cleanup
}

// runCLI executes the cobra root with args in the provided working directory.
// Output is discarded; any error is returned to the caller.
func runCLI(dir string, args ...string) error {
    c := newRootCmd()
    var out bytes.Buffer
    c.SetOut(&out)
    c.SetArgs(args)
    // change into workspace dir for deterministic relative path handling
    cwd, _ := os.Getwd()
    _ = os.Chdir(dir)
    defer func() { _ = os.Chdir(cwd) }()
    return c.Execute()
}

// prepareWorkspace initializes a fresh workspace in dir using `ami init --force`.
func prepareWorkspace(b *testing.B, dir string) {
    b.Helper()
    if err := runCLI(dir, "init", "--force", "--json"); err != nil {
        b.Fatalf("init workspace: %v", err)
    }
}

// prepopulateFor prepares additional state for specific subcommands to run meaningfully.
// For example, `mod sum` expects an existing ami.sum; we seed it via `mod update`.
func prepopulateFor(b *testing.B, dir string, name string) {
    b.Helper()
    switch name {
    case "mod sum", "mod list":
        // create ami.sum and populate cache from local workspace
        if err := runCLI(dir, "mod", "update", "--json"); err != nil {
            b.Fatalf("prepopulate (%s): mod update: %v", name, err)
        }
    case "pipeline visualize":
        // having an empty workspace is acceptable; optional source files can be added later
    }
}

// BenchmarkAMI_Subcommands measures typical runtime of core ami subcommands.
// Each sub-benchmark runs in its own sandbox to avoid cross-test interference.
func BenchmarkAMI_Subcommands(b *testing.B) {
    benches := []struct{
        name string
        args []string
    }{
        {name: "help", args: []string{"help"}},
        {name: "version", args: []string{"version"}},
        {name: "clean", args: []string{"clean"}},
        {name: "lint", args: []string{"lint", "--json"}},
        {name: "test", args: []string{"test", "--json"}},
        {name: "mod update", args: []string{"mod", "update", "--json"}},
        {name: "mod list", args: []string{"mod", "list", "--json"}},
        {name: "mod sum", args: []string{"mod", "sum", "--json"}},
        {name: "mod get", args: []string{"mod", "get", "./src", "--json"}},
        {name: "mod clean", args: []string{"mod", "clean", "--json"}},
        {name: "pipeline visualize", args: []string{"pipeline", "visualize", "--json", "--no-summary"}},
    }

    for _, bc := range benches {
        bc := bc // capture range var
        b.Run(bc.name, func(sb *testing.B) {
            dir, cleanup := benchSandbox(sb)
            defer cleanup()
            prepareWorkspace(sb, dir)
            prepopulateFor(sb, dir, bc.name)
            // warm-up: execute once to populate any caches without timing
            if err := runCLI(dir, bc.args...); err != nil {
                sb.Skipf("warmup %s failed (skipping benchmark): %v", bc.name, err)
                return
            }
            sb.ResetTimer()
            for i := 0; i < sb.N; i++ {
                if err := runCLI(dir, bc.args...); err != nil {
                    sb.Fatalf("run %s: %v", bc.name, err)
                }
            }
        })
    }
}
