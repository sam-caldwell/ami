package main

import (
    "os"
    "path/filepath"
    "strings"
    "time"

    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// scanSourceMergeSort performs a lightweight text scan for merge.Sort usage
// to flag missing field and invalid order arguments without requiring the
// full parser. It walks all .ami files under the given workspace-relative
// root and returns diagnostics with approximate positions.
func scanSourceMergeSort(dir, root string) []diag.Record {
    var out []diag.Record
    base := filepath.Clean(filepath.Join(dir, root))
    now := time.Now().UTC()
    _ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
        if err != nil || d.IsDir() || filepath.Ext(path) != ".ami" { return nil }
        b, err := os.ReadFile(path)
        if err != nil { return nil }
        s := string(b)
        low := strings.ToLower(s)
        idx := 0
        for {
            i := strings.Index(low[idx:], "merge.sort(")
            if i < 0 { break }
            i += idx
            // find closing paren from i
            j := strings.IndexByte(low[i:], ')')
            if j < 0 { break }
            j = i + j
            inner := strings.TrimSpace(low[i+len("merge.sort("):j])
            // compute approximate line/column
            line := 1
            col := 1
            for k := 0; k < i; k++ {
                if s[k] == '\n' { line++; col = 1 } else { col++ }
            }
            if inner == "" {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_MERGE_SORT_NO_FIELD", Message: "merge.Sort requires a field argument", File: path, Pos: &diag.Position{Line: line, Column: col, Offset: i}})
            } else {
                // parse up to two args
                parts := []string{}
                for _, p := range strings.Split(inner, ",") {
                    p = strings.TrimSpace(p)
                    if p != "" { parts = append(parts, p) }
                }
                if len(parts) >= 2 {
                    ord := parts[1]
                    if ord != "asc" && ord != "desc" {
                        out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MERGE_SORT_ORDER_INVALID", Message: "merge.Sort order must be 'asc' or 'desc'", File: path, Pos: &diag.Position{Line: line, Column: col, Offset: i}, Data: map[string]any{"order": parts[1]}})
                    }
                }
            }
            idx = j + 1
        }
        return nil
    })
    return out
}

