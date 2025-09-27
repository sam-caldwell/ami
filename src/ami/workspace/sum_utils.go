package workspace

import (
    "os"
    "path/filepath"
)

// UpdateManifestEntryFromCache computes the sha of ${cacheRoot}/name/version and records it in m.
func UpdateManifestEntryFromCache(m *Manifest, cacheRoot, name, version string) error {
    dir := filepath.Join(cacheRoot, name, version)
    sha, err := HashDir(dir)
    if err != nil { return err }
    m.Set(name, version, sha)
    return nil
}

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

// DefaultCacheRoot resolves the package cache root using AMI_PACKAGE_CACHE or default HOME path.
func DefaultCacheRoot() (string, error) {
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache != "" { return cache, nil }
    home, err := os.UserHomeDir()
    if err != nil { return "", err }
    return filepath.Join(home, ".ami", "pkg"), nil
}

