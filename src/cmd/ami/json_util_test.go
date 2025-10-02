package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_writeJSONFile(t *testing.T) {
    dir := t.TempDir()
    p := filepath.Join(dir, "x.json")
    if err := writeJSONFile(p, map[string]any{"a": 1}); err != nil { t.Fatal(err) }
    if _, err := os.Stat(p); err != nil { t.Fatal(err) }
}

