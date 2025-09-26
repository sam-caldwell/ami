package manifest

import "errors"

// Validate checks basic structural constraints on the manifest.
func (m *Manifest) Validate() error {
    if m == nil {
        return errors.New("nil manifest")
    }
    if m.Schema == "" {
        m.Schema = "ami.manifest/v1"
    }
    if m.Schema != "ami.manifest/v1" {
        return errors.New("invalid schema")
    }
    if m.Project.Name == "" || m.Project.Version == "" {
        return errors.New("project.name and project.version required")
    }
    // Do not auto-populate CreatedAt here to preserve deterministic writes.
    return nil
}

