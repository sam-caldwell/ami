package main

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
)

// Ensure runClean does not remove ami.sum in workspace root.
func TestClean_PreservesAmiSum(t *testing.T) {
    dir := filepath.Join("build", "test", "ami_clean", "preserve_sum")
    if err := os.MkdirAll(filepath.Join(dir, "build"), 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    // Seed ami.sum in root
    if err := os.WriteFile(filepath.Join(dir, "ami.sum"), []byte(`{"schema":"ami.sum/v1","packages":{}}`), 0o644); err != nil {
        t.Fatalf("write sum: %v", err)
    }
    var buf bytes.Buffer
    if err := runClean(&buf, dir, true); err != nil { t.Fatalf("runClean: %v", err) }
    if _, err := os.Stat(filepath.Join(dir, "ami.sum")); err != nil {
        t.Fatalf("ami.sum should persist across clean: %v", err)
    }
}

