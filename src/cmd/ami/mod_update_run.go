package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// types moved to mod_update_types.go

// runModUpdate copies local workspace packages to the cache and refreshes ami.sum.
// Remote resolution (git+ssh) and constraint solving are deferred to later phases.
func runModUpdate(out io.Writer, dir string, jsonOut bool) error {
    // Pre-check: audit current workspace/sum/cache to surface issues before update.
    auditRep, _ := workspace.AuditDependencies(dir) // best-effort; non-fatal

    var ws workspace.Workspace
    if err := ws.Load(filepath.Join(dir, "ami.workspace")); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "workspace not found or invalid"}) }
        return exit.New(exit.User, "workspace invalid: %v", err)
    }
    // Detect cache
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        if home, err := os.UserHomeDir(); err == nil {
            cache = filepath.Join(home, ".ami", "pkg")
        } else {
            cache = filepath.Join(os.TempDir(), "ami", "pkg")
        }
    }
    _ = os.MkdirAll(cache, 0o755)

    // Load existing ami.sum with workspace.Manifest (canonical nested format)
    sumPath := filepath.Join(dir, "ami.sum")
    var manifest workspace.Manifest
    if _, err := os.Stat(sumPath); err == nil {
        _ = manifest.Load(sumPath) // ignore error; fall back to empty
    }
    if manifest.Schema == "" { manifest.Schema = "ami.sum/v1" }
    if manifest.Packages == nil { manifest.Packages = map[string]map[string]string{} }

    var updated []modUpdateItem
    // Copy each workspace package with a root, name, version
    for _, e := range ws.Packages {
        p := e.Package
        if p.Name == "" || p.Version == "" || p.Root == "" { continue }
        src := filepath.Clean(filepath.Join(dir, p.Root))
        if st, err := os.Stat(src); err != nil || !st.IsDir() { continue }
        dst := filepath.Join(cache, p.Name, p.Version)
        // Remove and copy
        if err := os.RemoveAll(dst); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: fmt.Sprintf("remove failed: %v", err)}) }
            return exit.New(exit.IO, "remove: %v", err)
        }
        if err := copyDir(src, dst); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: fmt.Sprintf("copy failed: %v", err)}) }
            return exit.New(exit.IO, "copy: %v", err)
        }
        // Hash and update sum
        h, err := hashDir(dst)
        if err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "hash failed"}) }
            return exit.New(exit.IO, "hash failed: %v", err)
        }
        if manifest.Packages[p.Name] == nil { manifest.Packages[p.Name] = map[string]string{} }
        manifest.Packages[p.Name][p.Version] = h
        updated = append(updated, modUpdateItem{Name: p.Name, Version: p.Version, Path: dst})
    }

    // For remote requirements, compute best satisfying versions among existing manifest entries
    // (non-destructive; selection is reported, not enforced here).
    reqs, _ := workspace.CollectRemoteRequirements(&ws)
    var selected []modUpdateItem
    for _, r := range reqs {
        vers := manifest.Versions(r.Name)
        if len(vers) == 0 { continue }
        // include prereleases only if the constraint specifies a prerelease component
        includePre := strings.Contains(r.Constraint.Version, "-")
        if best, ok := workspace.HighestSatisfying(vers, r.Constraint, includePre); ok {
            selected = append(selected, modUpdateItem{Name: r.Name, Version: best})
        }
    }

    // Persist ami.sum using canonical Manifest writer
    if err := manifest.Save(sumPath); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "write ami.sum failed"}) }
        return exit.New(exit.IO, "write ami.sum: %v", err)
    }

    // Sort updated for deterministic output
    sort.Slice(updated, func(i, j int) bool {
        if updated[i].Name == updated[j].Name {
            return updated[i].Version < updated[j].Version
        }
        return updated[i].Name < updated[j].Name
    })

    if jsonOut {
        return json.NewEncoder(out).Encode(modUpdateResult{Updated: updated, Selected: selected, Message: "ok", Audit: embedAudit(auditRep)})
    }
    // Human-readable audit summary (non-fatal)
    if len(auditRep.ParseErrors) > 0 { _, _ = fmt.Fprintf(out, "audit: parse errors: %d\n", len(auditRep.ParseErrors)) }
    if !auditRep.SumFound { _, _ = fmt.Fprintln(out, "audit: ami.sum: not found") }
    if len(auditRep.MissingInSum) > 0 { _, _ = fmt.Fprintf(out, "audit: missing in sum: %s\n", joinCSV(auditRep.MissingInSum)) }
    if len(auditRep.Unsatisfied) > 0 { _, _ = fmt.Fprintf(out, "audit: unsatisfied: %s\n", joinCSV(auditRep.Unsatisfied)) }
    if len(auditRep.MissingInCache) > 0 { _, _ = fmt.Fprintf(out, "audit: missing in cache: %s\n", joinCSV(auditRep.MissingInCache)) }
    if len(auditRep.Mismatched) > 0 { _, _ = fmt.Fprintf(out, "audit: mismatched: %s\n", joinCSV(auditRep.Mismatched)) }
    for _, u := range updated {
        _, _ = fmt.Fprintf(out, "updated %s@%s -> %s\n", u.Name, u.Version, u.Path)
    }
    if len(selected) > 0 {
        for _, s := range selected {
            _, _ = fmt.Fprintf(out, "select %s@%s\n", s.Name, s.Version)
        }
    }
    return nil
}

// joinCSV joins a string slice with commas (no spaces) for compact summaries.
// joinCSV moved to mod_update_join.go

// modAuditEmbed mirrors key fields from AuditReport for JSON embedding in update result.
// modAuditEmbed and embedAudit moved to mod_update_audit.go
