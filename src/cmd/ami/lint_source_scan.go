package main

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
    "time"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceUnknown performs a lightweight scan for UNKNOWN_IDENT sentinel in .ami files
// under pkgRoot relative to dir, and emits W_UNKNOWN_IDENT with rudimentary positions.
func scanSourceUnknown(dir, pkgRoot string) []diag.Record {
    var diags []diag.Record
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() { return nil }
        if filepath.Ext(path) != ".ami" { return nil }
        f, err := os.Open(path)
        if err != nil { return nil }
        defer f.Close()
        now := time.Now().UTC()
        scanner := bufio.NewScanner(f)
        lineNo := 0
        for scanner.Scan() {
            lineNo++
            line := scanner.Text()
            if idx := strings.Index(line, "UNKNOWN_IDENT"); idx >= 0 {
                pos := &diag.Position{Line: lineNo, Column: idx + 1, Offset: 0}
                diags = append(diags, diag.Record{
                    Timestamp: now,
                    Level:     diag.Warn,
                    Code:      "W_UNKNOWN_IDENT",
                    Message:   "unknown identifier detected (scaffold)",
                    File:      path,
                    Pos:       pos,
                })
            }
        }
        return nil
    })
    return diags
}

