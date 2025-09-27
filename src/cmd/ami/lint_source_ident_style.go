package main

import (
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceIdentStyle scans .ami files for identifier naming style issues.
// Emits W_IDENT_UNDERSCORE for identifiers containing '_' (except single underscore).
func scanSourceIdentStyle(dir, pkgRoot string) []diag.Record {
    var diags []diag.Record
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
        b, err := os.ReadFile(path)
        if err != nil { return nil }
        f := &source.File{Name: path, Content: string(b)}
        s := scanner.New(f)
        now := time.Now().UTC()
        for {
            t := s.Next()
            if t.Kind == token.EOF { break }
            if t.Kind == token.Ident {
                if t.Lexeme != "_" && strings.Contains(t.Lexeme, "_") {
                    diags = append(diags, diag.Record{
                        Timestamp: now,
                        Level:     diag.Warn,
                        Code:      "W_IDENT_UNDERSCORE",
                        Message:   "identifier contains underscore; use camelCase or PascalCase",
                        File:      path,
                        Pos:       &diag.Position{Line: t.Pos.Line, Column: t.Pos.Column, Offset: t.Pos.Offset},
                    })
                }
            }
        }
        return nil
    })
    return diags
}

