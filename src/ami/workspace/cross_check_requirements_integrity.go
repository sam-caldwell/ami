package workspace

// CrossCheckRequirementsIntegrity filters manifest integrity results to the required dependencies.
// Returns sets of name@version keys missing in cache or mismatched vs recorded sha.
// A requirement is considered satisfied if any version in manifest satisfies the constraint; only
// those satisfying versions are considered when intersecting with integrity issues.
func CrossCheckRequirementsIntegrity(m *Manifest, reqs []Requirement) (missingInCache []string, mismatched []string, err error) {
    // Compute integrity across all entries.
    _, miss, mis, err := m.Validate()
    if err != nil { return nil, nil, err }
    // Build a set of satisfying keys for each requirement.
    satKeys := make(map[string]struct{})
    for _, r := range reqs {
        for _, v := range m.Versions(r.Name) {
            if Satisfies(v, r.Constraint) {
                satKeys[r.Name+"@"+v] = struct{}{}
            }
        }
    }
    // Intersect
    for _, k := range miss {
        if _, ok := satKeys[k]; ok { missingInCache = append(missingInCache, k) }
    }
    for _, k := range mis {
        if _, ok := satKeys[k]; ok { mismatched = append(mismatched, k) }
    }
    return missingInCache, mismatched, nil
}

