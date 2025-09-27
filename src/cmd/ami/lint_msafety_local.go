package main

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// analyzeMemorySafety performs lightweight, syntax-level checks to enforce AMI 2.3.2 memory safety:
// - Raw address-of operator '&' is not allowed.
// - Unary '*' is not a dereference; it's only allowed as a leading mutating assignment marker: "* name = expr".
// The analysis is tolerant and best-effort; it does not parse the language, only scans lines.
func analyzeMemorySafety(f *source.File) []diag.Record {
    var out []diag.Record
    if f == nil || f.Content == "" { return out }
    now := time.Now().UTC()
    // Scan line by line to produce stable positions.
    lines := splitLinesPreserve(f.Content)
    for i, line := range lines {
        // 1-based line number and simple column reporting.
        lineNo := i + 1
        // Pointer operator '&' is disallowed anywhere.
        if idx := strings.IndexByte(line, '&'); idx >= 0 {
            out = append(out, diag.Record{
                Timestamp: now,
                Level:     diag.Error,
                Code:      "E_PTR_UNSUPPORTED_SYNTAX",
                Message:   "address-of operator '&' is not allowed",
                File:      f.Name,
                Pos:       &diag.Position{Line: lineNo, Column: idx + 1, Offset: 0},
            })
        }
        // Unary '*' dereference is disallowed except for mutating assignment marker.
        t := strings.TrimLeft(line, " \t")
        if strings.HasPrefix(t, "*") {
            // Allowed pattern: "* name =" (star + space + identifier + optional spaces + '=')
            rest := strings.TrimSpace(t[1:])
            // Extract identifier token
            name := leadingIdent(rest)
            rem := strings.TrimLeft(rest[len(name):], " \t")
            if name == "" || !strings.HasPrefix(rem, "=") {
                out = append(out, diag.Record{
                    Timestamp: now,
                    Level:     diag.Error,
                    Code:      "E_MUT_BLOCK_UNSUPPORTED",
                    Message:   "unary '*' is not a dereference; only '* name = expr' is allowed",
                    File:      f.Name,
                    Pos:       &diag.Position{Line: lineNo, Column: strings.Index(line, "*") + 1, Offset: 0},
                })
            }
        }
    }
    return out
}

// splitLinesPreserve splits into lines, preserving empty trailing line behavior.
func splitLinesPreserve(s string) []string {
    // Use simple split that preserves content semantics for scanning.
    if s == "" { return []string{""} }
    // Normalize Windows newlines conservatively.
    s = strings.ReplaceAll(s, "\r\n", "\n")
    s = strings.ReplaceAll(s, "\r", "\n")
    return strings.Split(s, "\n")
}

// leadingIdent returns the leading identifier from s (letters/digits/underscore allowed in scan).
func leadingIdent(s string) string {
    i := 0
    for i < len(s) {
        c := s[i]
        if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
            i++
            continue
        }
        break
    }
    return s[:i]
}

