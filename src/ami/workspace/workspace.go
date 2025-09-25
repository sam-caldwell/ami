package workspace

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "runtime"
    "strings"

    yaml "gopkg.in/yaml.v3"
)

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

var osArchRe = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)
var semverRe = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

type Project struct {
    Name    string `yaml:"name"`
    Version string `yaml:"version"`
}

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

func (w *Workspace) Validate(repoRoot string) error {
    if w.Version == "" {
        return errors.New("workspace.version required (schema semver)")
    }
    // Project is required
    if w.Project.Name == "" {
        return errors.New("project.name required")
    }
    if w.Project.Version == "" || !semverRe.MatchString(w.Project.Version) {
        return errors.New("project.version must be SemVer (e.g., 0.0.1)")
    }
    // Compiler target default
    if strings.TrimSpace(w.Toolchain.Compiler.Target) == "" {
        w.Toolchain.Compiler.Target = "./build"
    }
    if filepath.IsAbs(w.Toolchain.Compiler.Target) {
        return errors.New("toolchain.compiler.target must be workspace-relative")
    }
    if strings.Contains(w.Toolchain.Compiler.Target, "..") {
        return errors.New("toolchain.compiler.target must not traverse outside workspace")
    }
    // Concurrency: int or "NUM_CPU"
    switch v := w.Toolchain.Compiler.Concurrency.(type) {
    case int, int64, int32:
        // ok
    case string:
        if strings.ToUpper(v) != "NUM_CPU" {
            return fmt.Errorf("invalid toolchain.compiler.concurrency: %q", v)
        }
    case nil:
        // default to NUM_CPU
        w.Toolchain.Compiler.Concurrency = "NUM_CPU"
    default:
        return errors.New("toolchain.compiler.concurrency must be integer or \"NUM_CPU\"")
    }
    // Env: if empty, default to darwin/arm64 for this phase
    if len(w.Toolchain.Compiler.Env) == 0 {
        w.Toolchain.Compiler.Env = []EnvTarget{{OS: "darwin/arm64"}}
    }
    // Validate env entries
    seen := map[string]bool{}
    uniq := make([]EnvTarget, 0, len(w.Toolchain.Compiler.Env))
    for _, e := range w.Toolchain.Compiler.Env {
        if !osArchRe.MatchString(e.OS) {
            return fmt.Errorf("invalid toolchain.compiler.env os: %q", e.OS)
        }
        if !seen[e.OS] {
            seen[e.OS] = true
            uniq = append(uniq, e)
        }
    }
    w.Toolchain.Compiler.Env = uniq
    _ = repoRoot
    return nil
}

// ResolveConcurrency returns effective concurrency as an integer.
func (w *Workspace) ResolveConcurrency() int {
    switch v := w.Toolchain.Compiler.Concurrency.(type) {
    case int:
        if v >= 1 { return v }
    case int64:
        if v >= 1 { return int(v) }
    case int32:
        if v >= 1 { return int(v) }
    case string:
        if strings.ToUpper(v) == "NUM_CPU" { return runtime.NumCPU() }
    }
    return 1
}
