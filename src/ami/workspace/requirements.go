package workspace

// Requirement represents a remote dependency declared by a workspace import.
// Local (./) imports are excluded from this representation.
type Requirement struct {
    Name       string
    Constraint Constraint
}

// CollectRemoteRequirements scans workspace packages and extracts remote import
// requirements (entries not starting with "./"). Parsing errors are skipped and
// returned in the errs slice for caller context. Order preserves package order.
func CollectRemoteRequirements(ws *Workspace) (reqs []Requirement, errs []error) {
    for _, e := range ws.Packages {
        p := e.Package
        NormalizeImports(&p)
        for _, ent := range p.Import {
            path, cstr := ParseImportEntry(ent)
            if path == "" || (len(path) >= 2 && path[:2] == "./") {
                continue // local or empty
            }
            // If no constraint, treat as latest for now.
            var c Constraint
            var err error
            if cstr == "" {
                c = Constraint{Op: OpLatest}
            } else {
                c, err = ParseConstraint(cstr)
                if err != nil { errs = append(errs, err); continue }
            }
            reqs = append(reqs, Requirement{Name: path, Constraint: c})
        }
    }
    return reqs, errs
}

// CrossCheckRequirements compares requirements against ami.sum manifest.
// Returns requirement names missing from sum and those present but without any
// version satisfying the constraint.
func CrossCheckRequirements(m *Manifest, reqs []Requirement) (missingInSum []string, unsatisfied []string) {
    for _, r := range reqs {
        versions := m.Versions(r.Name)
        if len(versions) == 0 {
            missingInSum = append(missingInSum, r.Name)
            continue
        }
        ok := false
        for _, v := range versions {
            if Satisfies(v, r.Constraint) { ok = true; break }
        }
        if !ok { unsatisfied = append(unsatisfied, r.Name) }
    }
    return missingInSum, unsatisfied
}

