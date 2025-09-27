package main

import (
    "os"
    "path/filepath"
    "time"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceLangNotGo emits W_LANG_NOT_GO for .go files under package roots to remind users
// that AMI is not Go. This is a gentle warning and can be silenced via config suppress rules.
func scanSourceLangNotGo(dir, pkgRoot string) []diag.Record {
    var diags []diag.Record
    root := filepath.Clean(filepath.Join(dir, pkgRoot))
    _ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil { return nil }
        if d.IsDir() { return nil }
        if filepath.Ext(path) != ".go" { return nil }
        diags = append(diags, diag.Record{
            Timestamp: time.Now().UTC(),
            Level:     diag.Warn,
            Code:      "W_LANG_NOT_GO",
            Message:   ".go file detected in AMI source tree",
            File:      path,
        })
        return nil
    })
    return diags
}

