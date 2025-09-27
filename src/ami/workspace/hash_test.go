package workspace

import (
    "os"
    "path/filepath"
    "testing"
)

func TestHashDir_DeterministicAndNonEmpty(t *testing.T) {
    dir := filepath.Join("build", "test", "workspace", "hashdet")
    _ = os.RemoveAll(dir)
    if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("A"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("B"), 0o644); err != nil { t.Fatalf("write: %v", err) }
    h1, err := HashDir(dir)
    if err != nil || len(h1) == 0 { t.Fatalf("hash1: %v %q", err, h1) }
    // Shuffling files should not change output; rewrite in different order
    if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("B"), 0o644); err != nil { t.Fatalf("rewrite: %v", err) }
    h2, err := HashDir(dir)
    if err != nil || h2 != h1 { t.Fatalf("hash2 mismatch: %v %q vs %q", err, h2, h1) }
}

