package workspace

import (
    "errors"
    "os"
    "path/filepath"
)

// AuditReport summarizes dependency status for a workspace against ami.sum and cache.
type AuditReport struct {
    // Requirements lists remote dependency requirements discovered across packages.
    Requirements []Requirement
    // MissingInSum lists requirement names not present in ami.sum.
    MissingInSum []string
    // Unsatisfied lists requirement names present in ami.sum but with no version satisfying the constraint.
    Unsatisfied []string
    // MissingInCache lists name@version entries (that satisfy constraints) that are missing from the cache.
    MissingInCache []string
    // Mismatched lists name@version entries (that satisfy constraints) with sha mismatches vs cache contents.
    Mismatched []string
    // ParseErrors captures import constraint parse errors as strings; requirements with parse errors are skipped.
    ParseErrors []string
    // SumFound indicates whether ami.sum existed and was parsed.
    SumFound bool
}

// AuditDependencies loads ami.workspace under dir, collects remote requirements, cross-checks them against
// ami.sum, and filters integrity issues to only versions that satisfy constraints. It returns an AuditReport
// suitable for consumption by CLI or other packages. This function performs no I/O beyond reading workspace
// and ami.sum; cache location is inferred by Manifest.Validate using AMI_PACKAGE_CACHE or HOME.
func AuditDependencies(dir string) (AuditReport, error) {
    var rep AuditReport
    // Load workspace
    wsPath := filepath.Join(dir, "ami.workspace")
    var ws Workspace
    if err := ws.Load(wsPath); err != nil { return rep, err }

    // Collect requirements
    reqs, perrs := CollectRemoteRequirements(&ws)
    rep.Requirements = reqs
    if len(perrs) > 0 {
        rep.ParseErrors = make([]string, 0, len(perrs))
        for _, e := range perrs { rep.ParseErrors = append(rep.ParseErrors, e.Error()) }
    }

    // Load manifest (optional)
    sumPath := filepath.Join(dir, "ami.sum")
    var m Manifest
    if _, err := os.Stat(sumPath); err == nil {
        if err := m.Load(sumPath); err != nil { return rep, err }
        rep.SumFound = true
    } else if !errors.Is(err, os.ErrNotExist) {
        return rep, err
    }

    // Cross-check against sum contents
    rep.MissingInSum, rep.Unsatisfied = CrossCheckRequirements(&m, reqs)

    // Cross-check integrity limited to satisfying versions
    miss, mis, err := CrossCheckRequirementsIntegrity(&m, reqs)
    if err != nil { return rep, err }
    rep.MissingInCache, rep.Mismatched = miss, mis
    return rep, nil
}

