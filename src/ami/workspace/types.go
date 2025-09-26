package workspace

type Workspace struct {
    Version   string    `yaml:"version"`
    Project   Project   `yaml:"project"`
    Toolchain Toolchain `yaml:"toolchain"`
    Packages  []any     `yaml:"packages"`
}

type Toolchain struct {
    Compiler Compiler `yaml:"compiler"`
    Linker   any      `yaml:"linker"`
    Linter   any      `yaml:"linter"`
}

type Compiler struct {
    Concurrency any         `yaml:"concurrency"`
    Target      string      `yaml:"target"`
    Env         []EnvTarget `yaml:"env"`
}

type EnvTarget struct {
    OS string `yaml:"os"`
}

type Project struct {
    Name    string `yaml:"name"`
    Version string `yaml:"version"`
}

