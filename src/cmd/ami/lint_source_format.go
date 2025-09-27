package main

import (
    "os"
    "path/filepath"
    "strings"
    "time"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceFormatting scans .ami files and emits formatting markers warnings:
// - W_FORMAT_TRAILING_WS: line ends with spaces or tabs
// - W_FORMAT_TAB_INDENT: leading indentation uses tabs
func scanSourceFormatting(dir, pkgRoot string) []diag.Record {
    var diags []diag.Record
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
        b, err := os.ReadFile(path)
        if err != nil { return nil }
        content := string(b)
        now := time.Now().UTC()
        // Normalize line endings
        content = strings.ReplaceAll(content, "\r\n", "\n")
        content = strings.ReplaceAll(content, "\r", "\n")
        lines := strings.Split(content, "\n")
        for i, ln := range lines {
            lineNo := i + 1
            // Tab indentation
            if len(ln) > 0 && ln[0] == '\t' {
                diags = append(diags, diag.Record{Timestamp: now, Level: diag.Info, Code: "W_FORMAT_TAB_INDENT", Message: "use spaces for indentation, not tabs", File: path, Pos: &diag.Position{Line: lineNo, Column: 1, Offset: 0}})
            }
            // Trailing whitespace
            if hasTrailingWhitespace(ln) {
                col := len(ln)
                diags = append(diags, diag.Record{Timestamp: now, Level: diag.Info, Code: "W_FORMAT_TRAILING_WS", Message: "remove trailing whitespace", File: path, Pos: &diag.Position{Line: lineNo, Column: col, Offset: 0}})
            }
        }
        return nil
    })
    return diags
}

func hasTrailingWhitespace(s string) bool {
    if s == "" { return false }
    // spaces or tabs at end
    i := len(s) - 1
    if s[i] == ' ' || s[i] == '\t' { return true }
    return false
}

