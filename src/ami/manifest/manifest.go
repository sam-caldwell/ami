package manifest

import (
    "encoding/json"
    "errors"
    "os"
    "sort"
)

type Manifest struct {
    Schema    string      `json:"schema"`
    Project   Project     `json:"project"`
    Packages  []Package   `json:"packages"`
    Artifacts []Artifact  `json:"artifacts"`
    Toolchain Toolchain   `json:"toolchain"`
    CreatedAt string      `json:"createdAt"`
}

type Project struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

type Package struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Digest  string `json:"digestSHA256"`
    Source  string `json:"source"`
}

type Artifact struct {
    Path   string `json:"path"`
    Kind   string `json:"kind"`
    Size   int64  `json:"size"`
    Sha256 string `json:"sha256"`
}

type Toolchain struct {
    AmiVersion string `json:"amiVersion"`
    GoVersion  string `json:"goVersion"`
}

func (m *Manifest) Validate() error {
    if m == nil { return errors.New("nil manifest") }
    if m.Schema == "" { m.Schema = "ami.manifest/v1" }
    if m.Schema != "ami.manifest/v1" { return errors.New("invalid schema") }
    if m.Project.Name == "" || m.Project.Version == "" { return errors.New("project.name and project.version required") }
    // Do not auto-populate CreatedAt here to preserve deterministic writes.
    return nil
}

// CrossCheckWithSumFile loads ami.sum from the given path and ensures each
// manifest package is present with the same digest. Returns an error on mismatch.
func (m *Manifest) CrossCheckWithSumFile(sumPath string) error {
    if m == nil { return errors.New("nil manifest") }
    b, err := os.ReadFile(sumPath)
    if err != nil { return err }
    // local minimal shape matching ami.sum
    var sum struct {
        Schema   string                         `json:"schema"`
        Packages map[string]map[string]string   `json:"packages"`
    }
    if err := json.Unmarshal(b, &sum); err != nil { return err }
    for _, p := range m.Packages {
        vers, ok := sum.Packages[p.Name]
        if !ok { return errors.New("ami.sum missing package: " + p.Name) }
        d, ok := vers[p.Version]
        if !ok { return errors.New("ami.sum missing version for package: " + p.Name) }
        if d != p.Digest { return errors.New("ami.sum digest mismatch for package: " + p.Name) }
    }
    return nil
}

func Load(path string) (*Manifest, error) {
    b, err := os.ReadFile(path)
    if err != nil { return nil, err }
    var m Manifest
    if err := json.Unmarshal(b, &m); err != nil { return nil, err }
    if err := m.Validate(); err != nil { return nil, err }
    return &m, nil
}

func Save(path string, m *Manifest) error {
    if err := m.Validate(); err != nil { return err }
    // deterministic ordering for packages and artifacts
    pkgs := make([]Package, len(m.Packages))
    copy(pkgs, m.Packages)
    sort.Slice(pkgs, func(i, j int) bool {
        if pkgs[i].Name == pkgs[j].Name { return pkgs[i].Version < pkgs[j].Version }
        return pkgs[i].Name < pkgs[j].Name
    })
    arts := make([]Artifact, len(m.Artifacts))
    copy(arts, m.Artifacts)
    sort.Slice(arts, func(i, j int) bool { return arts[i].Path < arts[j].Path })
    mc := *m
    mc.Packages = pkgs
    mc.Artifacts = arts
    b, err := json.MarshalIndent(&mc, "", "  ")
    if err != nil { return err }
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, b, 0644); err != nil { return err }
    return os.Rename(tmp, path)
}
