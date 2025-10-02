package workspace

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

