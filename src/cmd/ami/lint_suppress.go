package main

import (
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// applyConfigSuppress filters diagnostics using workspace toolchain.linter.suppress rules.
// Keys are path prefixes (workspace-relative); filtering is done on absolute cleaned paths.
func applyConfigSuppress(dir string, ws *workspace.Workspace, in []diag.Record) []diag.Record {
    if ws == nil || ws.Toolchain.Linter.Suppress == nil || len(ws.Toolchain.Linter.Suppress) == 0 {
        return in
    }
    // Precompute absolute prefix paths
    absPrefixes := make(map[string]map[string]bool)
    for rel, codes := range ws.Toolchain.Linter.Suppress {
        abs := filepath.Clean(filepath.Join(dir, rel))
        m := absPrefixes[abs]
        if m == nil { m = map[string]bool{}; absPrefixes[abs] = m }
        for _, c := range codes { if c != "" { m[c] = true } }
    }
    out := in[:0]
    for _, d := range in {
        file := d.File
        if file == "" { out = append(out, d); continue }
        suppressed := false
        for pfx, codes := range absPrefixes {
            if hasPathPrefix(file, pfx) && codes[d.Code] {
                suppressed = true
                break
            }
        }
        if !suppressed { out = append(out, d) }
    }
    return out
}

// hasPathPrefix reports whether path starts with prefix path segment-wise.
// hasPathPrefix moved to lint_suppress_prefix.go
