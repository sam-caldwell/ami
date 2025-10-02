package main

import (
    "strings"
    "time"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// analyzeMemorySafety performs lightweight, syntax-level checks to enforce AMI 2.3.2 memory safety.
func analyzeMemorySafety(f *source.File) []diag.Record {
    var out []diag.Record
    if f == nil || f.Content == "" { return out }
    now := time.Now().UTC()
    lines := splitLinesPreserve(f.Content)
    for i, line := range lines {
        lineNo := i + 1
        if idx := strings.IndexByte(line, '&'); idx >= 0 {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "address-of operator '&' is not allowed", File: f.Name, Pos: &diag.Position{Line: lineNo, Column: idx + 1, Offset: 0}})
        }
        t := strings.TrimLeft(line, " \t")
        if strings.HasPrefix(t, "*") {
            rest := strings.TrimSpace(t[1:])
            name := leadingIdent(rest)
            rem := strings.TrimLeft(rest[len(name):], " \t")
            if name == "" || !strings.HasPrefix(rem, "=") {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MUT_BLOCK_UNSUPPORTED", Message: "unary '*' is not a dereference; only '* name = expr' is allowed", File: f.Name, Pos: &diag.Position{Line: lineNo, Column: strings.Index(line, "*") + 1, Offset: 0}})
            }
        }
    }
    return out
}

