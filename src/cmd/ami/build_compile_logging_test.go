package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"

    "github.com/sam-caldwell/ami/src/ami/logging"
)

// Ensure compiler activity logs are written to build/debug/activity.log when verbose.
func TestBuild_CompilerActivity_LogsToActivityFile(t *testing.T) {
    dir := filepath.Join("build", "test", "build_compile_logging")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Minimal workspace and a source file
    if err := os.WriteFile(filepath.Join(dir, "ami.workspace"), []byte("version: 1.0.0\npackages:\n  - main:\n      name: app\n      version: 0.0.1\n      root: ./src\n      import: []\n"), 0o644); err != nil { t.Fatalf("write ws: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "src", "main.ami"), []byte("package app\n"), 0o644); err != nil { t.Fatalf("write src: %v", err) }

    // Install a verbose logger writing to build/debug/activity.log
    lg, _ := logging.New(logging.Options{JSON: true, Verbose: true, DebugDir: filepath.Join(dir, "build", "debug")})
    setRootLogger(lg)
    defer closeRootLogger()

    var buf bytes.Buffer
    if err := runBuild(&buf, dir, false, true); err != nil {
        // Build may fail for reasons outside this test; still check logging
    }
    // Verify activity.log exists and contains compiler events
    data, err := os.ReadFile(filepath.Join(dir, "build", "debug", "activity.log"))
    if err != nil { t.Fatalf("read activity.log: %v", err) }
    if !bytes.Contains(data, []byte("\"message\":\"compiler.start\"")) &&
        !bytes.Contains(data, []byte("\"message\":\"compiler.pkg.start\"")) {
        t.Fatalf("expected compiler activity logs; got: %s", string(data))
    }
}

