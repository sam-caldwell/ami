package workspace

// Workspace represents the root ami.workspace configuration file.
// It can be loaded from or saved to YAML.
type Workspace struct {
    Version   string     `yaml:"version"`
    Toolchain Toolchain  `yaml:"toolchain"`
    Packages  PackageList `yaml:"packages"`
}
