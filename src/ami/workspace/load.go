package workspace

import (
    "os"
    "path/filepath"

    yaml "gopkg.in/yaml.v3"
)

func Load(path string) (*Workspace, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var ws Workspace
    if err := yaml.Unmarshal(b, &ws); err != nil {
        return nil, err
    }
    if err := ws.Validate(filepath.Dir(path)); err != nil {
        return nil, err
    }
    return &ws, nil
}

