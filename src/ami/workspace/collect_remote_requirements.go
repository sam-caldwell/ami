package workspace

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

