package main

import (
    "path/filepath"
    "testing"
    "os"
)

func TestParseRuntimeCases_FindsCases(t *testing.T) {
    dir := t.TempDir()
    p := filepath.Join(dir, "x_test.ami")
    if err := os.WriteFile(p, []byte("#pragma test:case A\n#pragma test:runtime input={}\n"), 0o644); err != nil { t.Fatal(err) }
    cs, err := parseRuntimeCases(dir)
    if err != nil { t.Fatal(err) }
    if len(cs) != 1 || cs[0].Name != "A" { t.Fatalf("got: %#v", cs) }
}

