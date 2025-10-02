package workspace

import "path/filepath"

// UpdateManifestEntryFromCache computes the sha of ${cacheRoot}/name/version and records it in m.
func UpdateManifestEntryFromCache(m *Manifest, cacheRoot, name, version string) error {
    dir := filepath.Join(cacheRoot, name, version)
    sha, err := HashDir(dir)
    if err != nil { return err }
    m.Set(name, version, sha)
    return nil
}

