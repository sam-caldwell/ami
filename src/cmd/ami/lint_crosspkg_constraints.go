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
    // path -> list of constraints (all, including exact) for deeper checks
    all := map[string][]semver.Constraint{}
    // path -> list of non-exact constraints (for range intersection)
    ranges := map[string][]semver.Constraint{}
    for _, pe := range ws.Packages {
        for _, entry := range pe.Package.Import {
            path, constraint := splitImportConstraint(entry)
            if path == "" || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") { continue }
            if constraint == "" { continue }
            c, err := semver.ParseConstraint(constraint)
            if err != nil { continue }
            all[path] = append(all[path], c)
            if c.Latest { continue }
            if c.Op == "" { // exact
                vmap := vers[path]
                if vmap == nil { vmap = map[string]bool{}; vers[path] = vmap }
                v := c.Version
                vmap[v] = true
            } else {
                ranges[path] = append(ranges[path], c)
            }
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

    // Range incompatibility and single-version rule checks
    for p, list := range all {
        // 1) If there is at least one exact version, ensure it satisfies all constraints.
        var exacts []semver.Constraint
        var others []semver.Constraint
        for _, c := range list { if c.Op == "" { exacts = append(exacts, c) } else { others = append(others, c) } }
        if len(exacts) > 0 {
            // pick the first exact; conflicting exacts already flagged
            chosen := exacts[0]
            for _, c := range others { if !semver.Satisfies(chosen.Version, c) {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "E_IMPORT_CONSTRAINT", Message: "exact version " + chosen.Version + " does not satisfy constraint for " + p, File: "ami.workspace"})
                break } }
            continue
        }
        // 2) No exacts: intersect all range constraints; if empty → E_IMPORT_CONSTRAINT
        var ok bool
        var inter semver.Bound
        for i, c := range others {
            b, bok := semver.Bounds(c)
            if !bok { continue }
            if i == 0 { inter = b; ok = true; continue }
            if ok {
                inter, ok = semver.Intersect(inter, b)
                if !ok { break }
            }
        }
        if len(others) > 0 && !ok {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "E_IMPORT_CONSTRAINT", Message: "incompatible version constraints for " + p, File: "ami.workspace"})
            continue
        }
        // 3) Single-version rule (strict): emit warning when only ranges are present, asking to pin exact
        if len(others) > 0 {
            out = append(out, diag.Record{Timestamp: now, Level: diag.Warn, Code: "W_IMPORT_SINGLE_VERSION", Message: "no exact version pinned for " + p + "; strict mode requires a single pinned version", File: "ami.workspace"})
        }
    }
    return out
}

// constraintsConflict implements a conservative overlap check for two constraints.
// constraintsConflict moved to lint_crosspkg_constraints_conflict.go
