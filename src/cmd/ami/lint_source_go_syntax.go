package main

import (
    "os"
    "path/filepath"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceGoSyntax warns when a .ami file appears to contain Go source (e.g., leading `package`).
func scanSourceGoSyntax(dir, pkgRoot string) []diag.Record {
    var diags []diag.Record
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
        b, err := os.ReadFile(path)
        if err != nil { return nil }
        f := &source.File{Name: path, Content: string(b)}
        s := scanner.New(f)
        now := time.Now().UTC()
        // Examine the first non-comment token
        for {
            t := s.Next()
            if t.Kind == token.EOF { break }
            if t.Kind == token.LineComment || t.Kind == token.BlockComment { continue }
            if t.Kind == token.KwPackage {
                diags = append(diags, diag.Record{
                    Timestamp: now,
                    Level:     diag.Warn,
                    Code:      "W_GO_SYNTAX_DETECTED",
                    Message:   "file starts with 'package' — looks like Go syntax, not AMI",
                    File:      path,
                    Pos:       &diag.Position{Line: t.Pos.Line, Column: t.Pos.Column, Offset: t.Pos.Offset},
                })
            }
            break
        }
        return nil
    })
    return diags
}

