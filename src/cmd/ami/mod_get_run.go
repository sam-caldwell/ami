package main

import (
    "encoding/json"
    "fmt"
    "io"
    "io/fs"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

type modGetResult struct {
    Source  string `json:"source"`
    Name    string `json:"name"`
    Version string `json:"version"`
    Path    string `json:"path"`
    Message string `json:"message,omitempty"`
}

// runModGet fetches a module from a local path into the package cache
// and updates ami.sum in the workspace. git+ssh is planned for later milestones.
func runModGet(out io.Writer, dir string, src string, jsonOut bool) error {
    // Git sources (non-interactive) handled separately
    if strings.HasPrefix(src, "git+ssh://") || strings.HasPrefix(src, "file+git://") {
        return modGetGit(out, dir, src, jsonOut)
    }
    // Normalize and validate source is within workspace.
    wsPath := filepath.Join(dir, "ami.workspace")
    var ws workspace.Workspace
    if err := ws.Load(wsPath); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Message: "workspace not found or invalid"}) }
        return exit.New(exit.User, "workspace invalid: %v", err)
    }

    // Resolve absolute source path
    absSrc := src
    if !filepath.IsAbs(absSrc) {
        absSrc = filepath.Clean(filepath.Join(dir, src))
    }
    absDir := filepath.Clean(dir)
    // Ensure absSrc is within workspace dir
    if !strings.HasPrefix(absSrc+string(os.PathSeparator), absDir+string(os.PathSeparator)) {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Message: "source must be within workspace"}) }
        return exit.New(exit.User, "source must be within workspace")
    }
    // Identify package entry by matching Root.
    // Expect src to equal the package root or be under the root.
    var pkg *workspace.Package
    for i := range ws.Packages {
        root := ws.Packages[i].Package.Root
        if root == "" { continue }
        rabs := filepath.Clean(filepath.Join(dir, root))
        if absSrc == rabs || strings.HasPrefix(absSrc+string(os.PathSeparator), rabs+string(os.PathSeparator)) {
            pkg = &ws.Packages[i].Package
            break
        }
    }
    if pkg == nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Message: "source not declared in ami.workspace"}) }
        return exit.New(exit.User, "source not declared in ami.workspace")
    }
    if pkg.Name == "" || pkg.Version == "" {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Message: "package name/version missing in workspace"}) }
        return exit.New(exit.User, "package name/version missing in workspace")
    }

    // Determine cache path
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        home, _ := os.UserHomeDir()
        cache = filepath.Join(home, ".ami", "pkg")
    }
    dest := filepath.Join(cache, pkg.Name, pkg.Version)
    // Remove and copy
    if err := os.RemoveAll(dest); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: pkg.Name, Version: pkg.Version, Path: dest}) }
        return exit.New(exit.IO, "remove dest: %v", err)
    }
    if err := copyDir(absSrc, dest); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: pkg.Name, Version: pkg.Version, Path: dest}) }
        return exit.New(exit.IO, "copy failed: %v", err)
    }

    // Update ami.sum
    sumPath := filepath.Join(dir, "ami.sum")
    sum := map[string]any{"schema": "ami.sum/v1"}
    // read existing if present
    if b, err := os.ReadFile(sumPath); err == nil {
        var m map[string]any
        if json.Unmarshal(b, &m) == nil {
            sum = m
            if sum["schema"] != "ami.sum/v1" {
                sum = map[string]any{"schema": "ami.sum/v1"}
            }
        }
    }
    // compute hash of dest
    h, err := hashDir(dest)
    if err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: pkg.Name, Version: pkg.Version, Path: dest, Message: "hash failed"}) }
        return exit.New(exit.IO, "hash failed: %v", err)
    }
    // object form packages: name: {version, sha256}
    pkgs, _ := sum["packages"].(map[string]any)
    if pkgs == nil { pkgs = map[string]any{} }
    pkgs[pkg.Name] = map[string]any{"version": pkg.Version, "sha256": h}
    sum["packages"] = pkgs
    // write back
    if b, err := json.MarshalIndent(sum, "", "  "); err == nil {
        if err := os.WriteFile(sumPath, b, 0o644); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: pkg.Name, Version: pkg.Version, Path: dest, Message: "write ami.sum failed"}) }
            return exit.New(exit.IO, "write ami.sum: %v", err)
        }
    } else {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: pkg.Name, Version: pkg.Version, Path: dest, Message: "encode ami.sum failed"}) }
        return exit.New(exit.Internal, "encode ami.sum: %v", err)
    }

    if jsonOut {
        return json.NewEncoder(out).Encode(modGetResult{Source: src, Name: pkg.Name, Version: pkg.Version, Path: dest, Message: "ok"})
    }
    _, _ = fmt.Fprintf(out, "fetched %s@%s -> %s\n", pkg.Name, pkg.Version, dest)
    return nil
}

func copyDir(src, dst string) error {
    if err := os.MkdirAll(dst, 0o755); err != nil { return err }
    return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        // Skip VCS dirs
        if d.IsDir() && (d.Name() == ".git" || d.Name() == ".hg" || d.Name() == ".svn") {
            return filepath.SkipDir
        }
        rel, err := filepath.Rel(src, path)
        if err != nil { return err }
        target := filepath.Join(dst, rel)
        if d.IsDir() {
            return os.MkdirAll(target, 0o755)
        }
        b, err := os.ReadFile(path)
        if err != nil { return err }
        return os.WriteFile(target, b, 0o644)
    })
}

// modGetGit clones a git repository (git+ssh or file+git) at a specific tag
// into the package cache and updates ami.sum accordingly.
func modGetGit(out io.Writer, dir string, src string, jsonOut bool) error {
    // Parse URL and tag
    var repoURL, tag string
    if i := strings.LastIndex(src, "#"); i > 0 && i < len(src)-1 {
        repoURL, tag = src[:i], src[i+1:]
    } else {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Message: "missing tag"}) }
        return exit.New(exit.User, "git source must include #<tag>")
    }
    // Convert file+git to local path for git clone
    var cloneArg string
    if strings.HasPrefix(repoURL, "file+git://") {
        cloneArg = strings.TrimPrefix(repoURL, "file+git://")
        // ensure absolute path
        if cloneArg == "" || !filepath.IsAbs(cloneArg) {
            if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Message: "file+git requires absolute path"}) }
            return exit.New(exit.User, "file+git requires absolute path")
        }
    } else {
        // git+ssh stays as-is
        cloneArg = repoURL
    }
    // Derive name from repo path
    base := filepath.Base(cloneArg)
    if strings.HasSuffix(base, ".git") { base = strings.TrimSuffix(base, ".git") }
    name := base
    version := tag

    // Temp directory for clone
    tmp, err := os.MkdirTemp("", "ami-modget-")
    if err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Message: "tempdir failed"}) }
        return exit.New(exit.IO, "tempdir: %v", err)
    }
    defer os.RemoveAll(tmp)

    // Run git clone --depth 1 --branch <tag>
    env := os.Environ()
    env = append(env, "GIT_TERMINAL_PROMPT=0")
    env = append(env, "GIT_SSH_COMMAND=ssh -oBatchMode=yes -oStrictHostKeyChecking=no")
    cmd := exec.Command("git", "clone", "--depth", "1", "--branch", tag, cloneArg, tmp)
    cmd.Env = env
    if out != nil { /* we keep stdout/stderr quiet for determinism */ }
    if err := cmd.Run(); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Message: "git clone failed"}) }
        return exit.New(exit.IO, "git clone failed: %v", err)
    }

    // Determine cache path
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        home, _ := os.UserHomeDir()
        cache = filepath.Join(home, ".ami", "pkg")
    }
    dest := filepath.Join(cache, name, version)
    if err := os.RemoveAll(dest); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest}) }
        return exit.New(exit.IO, "remove dest: %v", err)
    }
    if err := copyDir(tmp, dest); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest}) }
        return exit.New(exit.IO, "copy failed: %v", err)
    }

    // Update ami.sum
    sumPath := filepath.Join(dir, "ami.sum")
    sum := map[string]any{"schema": "ami.sum/v1"}
    if b, err := os.ReadFile(sumPath); err == nil {
        var m map[string]any
        if json.Unmarshal(b, &m) == nil {
            sum = m
            if sum["schema"] != "ami.sum/v1" {
                sum = map[string]any{"schema": "ami.sum/v1"}
            }
        }
    }
    h, err := hashDir(dest)
    if err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest, Message: "hash failed"}) }
        return exit.New(exit.IO, "hash failed: %v", err)
    }
    pkgs, _ := sum["packages"].(map[string]any)
    if pkgs == nil { pkgs = map[string]any{} }
    pkgs[name] = map[string]any{"version": version, "sha256": h}
    sum["packages"] = pkgs
    if b, err := json.MarshalIndent(sum, "", "  "); err == nil {
        if err := os.WriteFile(sumPath, b, 0o644); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest, Message: "write ami.sum failed"}) }
            return exit.New(exit.IO, "write ami.sum: %v", err)
        }
    } else {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest, Message: "encode ami.sum failed"}) }
        return exit.New(exit.Internal, "encode ami.sum: %v", err)
    }

    if jsonOut {
        return json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest, Message: "ok"})
    }
    _, _ = fmt.Fprintf(out, "fetched %s@%s -> %s\n", name, version, dest)
    return nil
}
