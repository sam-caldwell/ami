package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_scanSourceGoSyntax_detectsPackage(t *testing.T) {
    dir := t.TempDir(); root := "pkg"; base := filepath.Join(dir, root)
    _ = os.MkdirAll(base, 0o755)
    _ = os.WriteFile(filepath.Join(base, "x.ami"), []byte("package x\n"), 0o644)
    ds := scanSourceGoSyntax(dir, root)
    if len(ds) == 0 { t.Fatal("expected diagnostic") }
}

