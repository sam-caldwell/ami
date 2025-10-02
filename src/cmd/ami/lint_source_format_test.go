package main

import (
    "os"
    "path/filepath"
    "testing"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

func Test_scanSourceFormatting_FindsMarkers(t *testing.T) {
    dir := t.TempDir()
    // place under pkgRoot subdir
    pkgRoot := "pkg"
    root := filepath.Join(dir, pkgRoot)
    if err := os.MkdirAll(root, 0o755); err != nil { t.Fatal(err) }
    // file with tab indent and trailing spaces
    content := "\tline with tab\nno-trailing\nline with trailing \t\n"
    if err := os.WriteFile(filepath.Join(root, "x.ami"), []byte(content), 0o644); err != nil { t.Fatal(err) }

    diags := scanSourceFormatting(dir, pkgRoot)
    if len(diags) == 0 { t.Fatal("expected at least one diagnostic") }
    // ensure codes set contains the two codes we emit
    want := map[string]bool{"W_FORMAT_TAB_INDENT": false, "W_FORMAT_TRAILING_WS": false}
    for _, d := range diags {
        if d.Level != diag.Info { t.Fatalf("unexpected level: %v", d.Level) }
        if _, ok := want[d.Code]; ok { want[d.Code] = true }
    }
    for code, seen := range want { if !seen { t.Fatalf("missing code: %s", code) } }
}

