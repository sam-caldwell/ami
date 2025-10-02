package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

// modGetGit clones a git repository (git+ssh or file+git) at a specific tag
// into the package cache and updates ami.sum accordingly.
func modGetGit(out io.Writer, dir string, src string, jsonOut bool) error {
    // Parse URL and optional tag
    var repoURL, tag string
    if i := strings.LastIndex(src, "#"); i > 0 && i < len(src)-1 {
        repoURL, tag = src[:i], src[i+1:]
    } else {
        repoURL = src
        tag = "" // will be resolved to highest non-prerelease tag
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

    // If tag omitted, resolve highest non-prerelease SemVer tag
    if version == "" {
        tgs, err := listGitTags(cloneArg)
        if err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Message: "list tags failed"}) }
            return exit.New(exit.Network, "list tags: %v", err)
        }
        version, err = selectHighestSemver(tgs, false)
        if err != nil || version == "" {
            if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Message: "no semver tags found"}) }
            return exit.New(exit.User, "no semver tags found")
        }
    }

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
    env = append(env, "GIT_SSH_COMMAND=ssh -oBatchMode=yes -oStrictHostKeyChecking=no -oConnectTimeout=2")
    cmd := exec.Command("git", "clone", "--depth", "1", "--branch", tag, cloneArg, tmp)
    cmd.Env = env
    if out != nil { /* keep stdout/stderr quiet for determinism */ }
    if err := cmd.Run(); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Message: "git clone failed"}) }
        return exit.New(exit.Network, "git clone failed: %v", err)
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
    // Compute directory hash for cache verification in this phase
    h, err := hashDir(dest)
    if err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modGetResult{Source: src, Name: name, Version: version, Path: dest, Message: "hash failed"}) }
        return exit.New(exit.IO, "hash failed: %v", err)
    }
    // also compute commit digest for resolution trace (non-breaking addition)
    commitDigest, _ := computeCommitDigest(tmp, version)
    pkgs, _ := sum["packages"].(map[string]any)
    if pkgs == nil { pkgs = map[string]any{} }
    entry := map[string]any{"version": version, "sha256": h}
    if commitDigest != "" { entry["commit"] = commitDigest }
    pkgs[name] = entry
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
