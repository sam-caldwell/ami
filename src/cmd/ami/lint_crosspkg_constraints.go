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
    // path -> list of constraints (non-exact) for overlap checks
    ranges := map[string][]semver.Constraint{}
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

    // Build range/other constraint lists and detect prerelease conflicts
    // path -> flags
    prereleaseSeen := map[string]bool{}
    nonPreSeen := map[string]bool{}
    // Re-scan including all constraints
    for _, pe := range ws.Packages {
        for _, entry := range pe.Package.Import {
            path, constraint := splitImportConstraint(entry)
            if path == "" || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") { continue }
            if constraint == "" { continue }
            c, err := semver.ParseConstraint(constraint)
            if err != nil { continue }
            if c.Latest { continue }
            if c.Op == "" {
                // exact: mark prerelease or not
                if strings.Contains(c.Version, "-") { prereleaseSeen[path] = true } else { nonPreSeen[path] = true }
            } else {
                ranges[path] = append(ranges[path], c)
                // consider range anchors without prerelease as non-pre constraints
                if !strings.Contains(c.Version, "-") { nonPreSeen[path] = true } else { prereleaseSeen[path] = true }
            }
        }
    }
    for p := range prereleaseSeen {
        if nonPreSeen[p] {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "E_IMPORT_PRERELEASE_FORBIDDEN", Message: "prerelease exact import used alongside non-prerelease constraints for " + p, File: "ami.workspace"})
        }
    }

    // Conservative range incompatibility checks
    for p, list := range ranges {
        // pairwise check
        for i := 0; i < len(list); i++ {
            for j := i+1; j < len(list); j++ {
                if constraintsConflict(list[i], list[j]) {
                    out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "E_IMPORT_CONSTRAINT", Message: "incompatible version constraints for " + p, File: "ami.workspace"})
                    break
                }
            }
        }
    }
    return out
}

// constraintsConflict implements a conservative overlap check for two constraints.
func constraintsConflict(a, b semver.Constraint) bool {
    // exact handled elsewhere
    if a.Op == "" || b.Op == "" { return false }
    // ^ vs ^ with different major → conflict
    if a.Op == "^" && b.Op == "^" {
        va, _ := semver.ParseVersion(a.Version)
        vb, _ := semver.ParseVersion(b.Version)
        return va.Major != vb.Major
    }
    // ~ vs ~ with different major/minor → conflict
    if a.Op == "~" && b.Op == "~" {
        va, _ := semver.ParseVersion(a.Version)
        vb, _ := semver.ParseVersion(b.Version)
        return va.Major != vb.Major || va.Minor != vb.Minor
    }
    // ^ vs ~ conflict when majors differ
    if (a.Op == "^" && b.Op == "~") || (a.Op == "~" && b.Op == "^") {
        va, _ := semver.ParseVersion(a.Version)
        vb, _ := semver.ParseVersion(b.Version)
        return va.Major != vb.Major
    }
    // >= ranges: conservative, assume overlap
    return false
}
