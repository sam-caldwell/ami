package workspace

import "sort"

// Manifest represents the ami.sum file in memory.
// It normalizes both supported shapes into a nested map:
//   Packages[packageName][version] = sha256
type Manifest struct {
    Schema   string
    Packages map[string]map[string]string
}

// Has reports whether name@version exists in the manifest with any sha.
func (m *Manifest) Has(name, version string) bool {
    if m.Packages == nil { return false }
    v, ok := m.Packages[name]
    if !ok { return false }
    _, ok = v[version]
    return ok
}

// Set records name@version=sha in the manifest, creating maps as needed.
func (m *Manifest) Set(name, version, sha string) {
    if m.Packages == nil { m.Packages = map[string]map[string]string{} }
    if m.Packages[name] == nil { m.Packages[name] = map[string]string{} }
    m.Packages[name][version] = sha
}

// Versions returns a sorted list of versions for a package name.
func (m *Manifest) Versions(name string) []string {
    mm := m.Packages[name]
    if mm == nil { return nil }
    out := make([]string, 0, len(mm))
    for ver := range mm { out = append(out, ver) }
    sort.Strings(out)
    return out
}

