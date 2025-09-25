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
    if !semverRe.MatchString(w.Version) {
        return errors.New("workspace.version must be SemVer (e.g., 1.0.0)")
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
    // Concurrency: int>=1 or "NUM_CPU"
    switch v := w.Toolchain.Compiler.Concurrency.(type) {
    case int:
        if v < 1 { return errors.New("toolchain.compiler.concurrency must be >=1") }
    case int64:
        if v < 1 { return errors.New("toolchain.compiler.concurrency must be >=1") }
    case int32:
        if v < 1 { return errors.New("toolchain.compiler.concurrency must be >=1") }
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

    // Linker and Linter must be objects (reserved for future keys)
    if w.Toolchain.Linker == nil {
        return errors.New("toolchain.linker must be an object (use {} if empty)")
    }
    switch w.Toolchain.Linker.(type) {
    case map[string]any:
        // ok
    default:
        return errors.New("toolchain.linker must be an object")
    }
    if w.Toolchain.Linter == nil {
        return errors.New("toolchain.linter must be an object (use {} if empty)")
    }
    switch w.Toolchain.Linter.(type) {
    case map[string]any:
        // ok
    default:
        return errors.New("toolchain.linter must be an object")
    }

    // Validate package import constraints if present
    if err := w.validatePackageImports(); err != nil { return err }
    _ = repoRoot
    return nil
}

// validatePackageImports inspects packages[*].<pkg>.import (if present) and
// validates entries as "<path> [constraint]", where constraint is one of:
//  - exact semver (with optional leading 'v'), e.g., v1.2.3 or 1.2.3
//  - ^<semver> (caret), ~<semver> (tilde)
//  - > <semver>, >=<semver>
//  - ==latest (macro)
func (w *Workspace) validatePackageImports() error {
    for _, p := range w.Packages {
        m, ok := p.(map[string]any)
        if !ok { continue }
        for _, v := range m {
            pm, ok := v.(map[string]any)
            if !ok { continue }
            imp, ok := pm["import"]
            if !ok || imp == nil { continue }
            lst, ok := imp.([]any)
            if !ok { return errors.New("packages.import must be a sequence of strings") }
            for _, item := range lst {
                s, ok := item.(string)
                if !ok { return errors.New("packages.import items must be strings") }
                fields := strings.Fields(s)
                if len(fields) == 0 { return errors.New("packages.import contains empty entry") }
                if len(fields) > 2 { return fmt.Errorf("invalid import entry (too many tokens): %q", s) }
                // path is fields[0], constraint optional
                if len(fields) == 2 {
                    cons := strings.ReplaceAll(fields[1], " ", "")
                    if !isValidConstraint(cons) {
                        return fmt.Errorf("invalid version constraint: %q", fields[1])
                    }
                }
            }
        }
    }
    return nil
}

func isValidConstraint(s string) bool {
    if s == "" || s == "==latest" { return true }
    if semverRe.MatchString(s) { return true }
    if strings.HasPrefix(s, "^") { return semverRe.MatchString(strings.TrimPrefix(s, "^")) }
    if strings.HasPrefix(s, "~") { return semverRe.MatchString(strings.TrimPrefix(s, "~")) }
    if strings.HasPrefix(s, ">=") { return semverRe.MatchString(strings.TrimPrefix(s, ">=")) }
    if strings.HasPrefix(s, ">") { return semverRe.MatchString(strings.TrimPrefix(s, ">")) }
    return false
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
