package manifest

import (
    "encoding/json"
    "errors"
    "os"
    "time"
)

type Manifest struct {
    Schema   string      `json:"schema"`
    Project  Project     `json:"project"`
    Packages []Package   `json:"packages"`
    Artifacts []Artifact `json:"artifacts"`
    Toolchain Toolchain  `json:"toolchain"`
    CreatedAt string     `json:"createdAt"`
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
    Path string `json:"path"`
    Kind string `json:"kind"`
    Size int64  `json:"size"`
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
    if m.CreatedAt == "" { m.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano) }
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
    b, err := json.MarshalIndent(m, "", "  ")
    if err != nil { return err }
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, b, 0644); err != nil { return err }
    return os.Rename(tmp, path)
}
