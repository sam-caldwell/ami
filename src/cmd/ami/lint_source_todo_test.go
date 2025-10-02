package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_scanSourceTodos_flagsMarkers(t *testing.T) {
    dir := t.TempDir(); root := "pkg"; base := filepath.Join(dir, root)
    _ = os.MkdirAll(base, 0o755)
    _ = os.WriteFile(filepath.Join(base, "x.ami"), []byte("TODO\nFIXME"), 0o644)
    ds := scanSourceTodos(dir, root)
    if len(ds) < 2 { t.Fatal("expected TODO and FIXME") }
}

