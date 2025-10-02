package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_scanSourceLangNotGo_flagsGoFiles(t *testing.T) {
    dir := t.TempDir(); root := "pkg"; base := filepath.Join(dir, root)
    _ = os.MkdirAll(base, 0o755)
    _ = os.WriteFile(filepath.Join(base, "x.go"), []byte("package x"), 0o644)
    ds := scanSourceLangNotGo(dir, root)
    if len(ds) == 0 { t.Fatal("expected a diagnostic") }
}

