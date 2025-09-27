package workspace

// Toolchain groups compiler, linker, and linter configuration.
type Toolchain struct {
    Compiler Compiler `yaml:"compiler"`
    Linker   Linker   `yaml:"linker"`
    Linter   Linter   `yaml:"linter"`
}

