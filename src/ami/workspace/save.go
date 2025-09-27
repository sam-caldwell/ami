package workspace

import (
    "bytes"
    "os"
    "gopkg.in/yaml.v3"
)

// Save writes the workspace as YAML to the given path.
func (w Workspace) Save(path string) error {
    var buf bytes.Buffer
    // Add YAML document marker for readability consistency with SPEC.
    buf.WriteString("---\n")
    enc := yaml.NewEncoder(&buf)
    enc.SetIndent(2)
    if err := enc.Encode(w); err != nil {
        _ = enc.Close()
        return err
    }
    if err := enc.Close(); err != nil {
        return err
    }
    return os.WriteFile(path, buf.Bytes(), 0o644)
}

