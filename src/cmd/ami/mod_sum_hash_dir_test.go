package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_hashDir_DeterministicAndContentSensitive(t *testing.T) {
    dir := t.TempDir()
    // create files out of order
    if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("world"), 0o644); err != nil { t.Fatal(err) }
    if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil { t.Fatal(err) }

    h1, err := hashDir(dir)
    if err != nil { t.Fatalf("hash1 error: %v", err) }
    h2, err := hashDir(dir)
    if err != nil { t.Fatalf("hash2 error: %v", err) }
    if h1 != h2 { t.Fatalf("expected deterministic hash, got %q vs %q", h1, h2) }

    // change content -> hash must change
    if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("HELLO"), 0o644); err != nil { t.Fatal(err) }
    h3, err := hashDir(dir)
    if err != nil { t.Fatalf("hash3 error: %v", err) }
    if h3 == h1 { t.Fatalf("expected different hash after content change") }
}

