package workspace

// Package describes a workspace package entry.
type Package struct {
    Name    string   `yaml:"name"`
    Version string   `yaml:"version"`
    Root    string   `yaml:"root"`
    Import  []string `yaml:"import"`
}

