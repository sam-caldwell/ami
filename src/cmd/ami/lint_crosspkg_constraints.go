package main

import (
    "strings"
    "time"

    "github.com/sam-caldwell/ami/src/ami/semver"
    "github.com/sam-caldwell/ami/src/ami/workspace"
    diag "github.com/sam-caldwell/ami/src/schemas/diag"
)

// lintCrossPackageConstraints checks for conflicting exact versions for the same import path across packages.
// Emits E_IMPORT_CONSTRAINT_MULTI when multiple exact versions differ. Only analyzes exact versions.
func lintCrossPackageConstraints(ws *workspace.Workspace) []diag.Record {
    var out []diag.Record
    if ws == nil { return out }
    now := time.Now().UTC()
    // path -> set of exact versions (normalized)
    vers := map[string]map[string]bool{}
    for _, pe := range ws.Packages {
        for _, entry := range pe.Package.Import {
            path, constraint := splitImportConstraint(entry)
            if path == "" || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") { continue }
            if constraint == "" { continue }
            c, err := semver.ParseConstraint(constraint)
            if err != nil || c.Latest || c.Op != "" { continue }
            vmap := vers[path]
            if vmap == nil { vmap = map[string]bool{}; vers[path] = vmap }
            v := c.Version
            vmap[v] = true
        }
    }
    for p, vset := range vers {
        if len(vset) <= 1 { continue }
        // conflicting exact versions
        out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "E_IMPORT_CONSTRAINT_MULTI", Message: "conflicting exact versions for " + p, File: "ami.workspace"})
    }
    return out
}

