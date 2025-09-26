package manifest

import (
    "encoding/json"
    "os"
    "sort"
)

// Load reads a manifest from a JSON file and validates it.
func Load(path string) (*Manifest, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var m Manifest
    if err := json.Unmarshal(b, &m); err != nil {
        return nil, err
    }
    if err := m.Validate(); err != nil {
        return nil, err
    }
    return &m, nil
}

// Save writes the manifest to a JSON file with deterministic ordering.
func Save(path string, m *Manifest) error {
    if err := m.Validate(); err != nil {
        return err
    }
    // deterministic ordering for packages and artifacts
    pkgs := make([]Package, len(m.Packages))
    copy(pkgs, m.Packages)
    sort.Slice(pkgs, func(i, j int) bool {
        if pkgs[i].Name == pkgs[j].Name {
            return pkgs[i].Version < pkgs[j].Version
        }
        return pkgs[i].Name < pkgs[j].Name
    })
    arts := make([]Artifact, len(m.Artifacts))
    copy(arts, m.Artifacts)
    sort.Slice(arts, func(i, j int) bool { return arts[i].Path < arts[j].Path })
    mc := *m
    mc.Packages = pkgs
    mc.Artifacts = arts
    b, err := json.MarshalIndent(&mc, "", "  ")
    if err != nil {
        return err
    }
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, b, 0644); err != nil {
        return err
    }
    return os.Rename(tmp, path)
}

