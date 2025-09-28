package workspace

// Compiler holds compiler-specific options.
type Compiler struct {
    Concurrency string   `yaml:"concurrency"`
    Target      string   `yaml:"target"`
    Env         []string `yaml:"env"`
    Backend     string   `yaml:"backend"`
    Options     []string `yaml:"options"`
}
