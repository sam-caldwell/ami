package workspace

import (
    "fmt"
    "path/filepath"
    "regexp"
    "strings"
)

var (
    reSemVer   = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)
    reOsArch   = regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)
    rePkgName  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9-_.]*$`)
)

// Validate checks the Workspace for schema-level constraints required by build.
// It returns a slice of human-readable error messages (empty when valid).
func (w Workspace) Validate() []string {
    var errs []string
    // version: SemVer required
    if !reSemVer.MatchString(strings.TrimSpace(w.Version)) {
        errs = append(errs, "version must be SemVer MAJOR.MINOR.PATCH (optionally with prerelease)")
    }
    // toolchain.compiler.concurrency: int>=1 or NUM_CPU
    c := strings.TrimSpace(w.Toolchain.Compiler.Concurrency)
    if c == "" {
        // treat as default NUM_CPU; no error
    } else if c != "NUM_CPU" {
        ok := true
        for _, r := range c {
            if r < '0' || r > '9' { ok = false; break }
        }
        if !ok {
            errs = append(errs, "toolchain.compiler.concurrency must be an integer >=1 or NUM_CPU")
        } else if c == "0" {
            errs = append(errs, "toolchain.compiler.concurrency must be >=1")
        }
    }
    // toolchain.compiler.target: workspace-relative, not absolute, no parent traversal
    tgt := strings.TrimSpace(w.Toolchain.Compiler.Target)
    if tgt == "" { tgt = "./build" }
    if filepath.IsAbs(tgt) {
        errs = append(errs, "toolchain.compiler.target must be workspace-relative (not absolute)")
    }
    clean := filepath.Clean(tgt)
    if strings.HasPrefix(clean, "..") || strings.Contains(clean, string(filepath.Separator)+"..") {
        errs = append(errs, "toolchain.compiler.target must not traverse outside workspace")
    }
    // toolchain.compiler.env: each entry os/arch
    for i, e := range w.Toolchain.Compiler.Env {
        if !reOsArch.MatchString(strings.TrimSpace(e)) {
            errs = append(errs, fmt.Sprintf("toolchain.compiler.env[%d] must be os/arch", i))
        }
    }
    // packages: basic shape (name, version, root) and name format; version semver
    // main package is not strictly required at this stage; build may target other flows later.
    for _, p := range w.Packages {
        if strings.TrimSpace(p.Package.Name) == "" {
            errs = append(errs, fmt.Sprintf("packages.%s.name missing", p.Key))
        } else if !rePkgName.MatchString(p.Package.Name) {
            errs = append(errs, fmt.Sprintf("packages.%s.name has invalid characters", p.Key))
        }
        if strings.TrimSpace(p.Package.Version) == "" || !reSemVer.MatchString(p.Package.Version) {
            errs = append(errs, fmt.Sprintf("packages.%s.version must be SemVer", p.Key))
        }
        r := strings.TrimSpace(p.Package.Root)
        if r == "" {
            errs = append(errs, fmt.Sprintf("packages.%s.root missing", p.Key))
        } else if filepath.IsAbs(r) || strings.HasPrefix(filepath.Clean(r), "..") {
            errs = append(errs, fmt.Sprintf("packages.%s.root must be workspace-relative and not escape", p.Key))
        }
    }
    // packages.main requirement relaxed for current phase.
    return errs
}
