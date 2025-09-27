package main

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"
    "time"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceTodos scans .ami files for TODO/FIXME markers and emits warnings.
// Codes: W_TODO, W_FIXME. Positions point to the start of the marker.
func scanSourceTodos(dir, pkgRoot string) []diag.Record {
    var diags []diag.Record
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
        f, err := os.Open(path)
        if err != nil { return nil }
        defer f.Close()
        now := time.Now().UTC()
        s := bufio.NewScanner(f)
        lineNo := 0
        for s.Scan() {
            lineNo++
            line := s.Text()
            if idx := strings.Index(line, "TODO"); idx >= 0 {
                diags = append(diags, diag.Record{
                    Timestamp: now,
                    Level:     diag.Warn,
                    Code:      "W_TODO",
                    Message:   "TODO found in source",
                    File:      path,
                    Pos:       &diag.Position{Line: lineNo, Column: idx + 1, Offset: 0},
                })
            }
            if idx := strings.Index(line, "FIXME"); idx >= 0 {
                diags = append(diags, diag.Record{
                    Timestamp: now,
                    Level:     diag.Warn,
                    Code:      "W_FIXME",
                    Message:   "FIXME found in source",
                    File:      path,
                    Pos:       &diag.Position{Line: lineNo, Column: idx + 1, Offset: 0},
                })
            }
        }
        return nil
    })
    return diags
}

