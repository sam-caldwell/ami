package workspace

import (
    "errors"
    "fmt"
    "path/filepath"
    "strings"
)

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
        if v < 1 {
            return errors.New("toolchain.compiler.concurrency must be >=1")
        }
    case int64:
        if v < 1 {
            return errors.New("toolchain.compiler.concurrency must be >=1")
        }
    case int32:
        if v < 1 {
            return errors.New("toolchain.compiler.concurrency must be >=1")
        }
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
    if err := w.validatePackageImports(); err != nil {
        return err
    }
    _ = repoRoot
    return nil
}

