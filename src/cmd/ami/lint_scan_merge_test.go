package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_scanSourceMergeSort_reportsIssues(t *testing.T) {
    dir := t.TempDir(); root := "pkg"; base := filepath.Join(dir, root)
    _ = os.MkdirAll(base, 0o755)
    _ = os.WriteFile(filepath.Join(base, "x.ami"), []byte("merge.Sort(,bad)\nmerge.Sort(field,down)"), 0o644)
    ds := scanSourceMergeSort(dir, root)
    if len(ds) < 1 { t.Fatal("expected diagnostics") }
}

