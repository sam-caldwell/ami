package main

import (
    "os"
    "path/filepath"
    "testing"
)

func Test_scanPragmas_basic(t *testing.T) {
    dir := t.TempDir()
    root := "pkg"
    base := filepath.Join(dir, root)
    if err := os.MkdirAll(base, 0o755); err != nil { t.Fatal(err) }
    src := "#pragma lint:disable A,B\n#pragma lint:enable A\n"
    if err := os.WriteFile(filepath.Join(base, "x.ami"), []byte(src), 0o644); err != nil { t.Fatal(err) }
    m := scanPragmas(dir, root)
    if !m[filepath.Join(base, "x.ami")]["B"] { t.Fatalf("expected B disabled: %#v", m) }
}

