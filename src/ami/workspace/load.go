package workspace

import (
    "os"
    "gopkg.in/yaml.v3"
)

// Load reads YAML from path into the workspace receiver.
func (w *Workspace) Load(path string) error {
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    return yaml.Unmarshal(b, w)
}

