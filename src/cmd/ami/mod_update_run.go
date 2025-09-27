package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "sort"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

type modUpdateItem struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Path    string `json:"path"`
}

type modUpdateResult struct {
    Updated []modUpdateItem `json:"updated"`
    Message string          `json:"message,omitempty"`
}

// runModUpdate copies local workspace packages to the cache and refreshes ami.sum.
// Remote resolution (git+ssh) and constraint solving are deferred to later phases.
func runModUpdate(out io.Writer, dir string, jsonOut bool) error {
    var ws workspace.Workspace
    if err := ws.Load(filepath.Join(dir, "ami.workspace")); err != nil {
        if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "workspace not found or invalid"}) }
        return exit.New(exit.User, "workspace invalid: %v", err)
    }
    // Detect cache
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        if home, err := os.UserHomeDir(); err == nil {
            cache = filepath.Join(home, ".ami", "pkg")
        } else {
            cache = filepath.Join(os.TempDir(), "ami", "pkg")
        }
    }
    _ = os.MkdirAll(cache, 0o755)

    // Load existing ami.sum (object form)
    sumPath := filepath.Join(dir, "ami.sum")
    sum := map[string]any{"schema": "ami.sum/v1"}
    if b, err := os.ReadFile(sumPath); err == nil {
        var m map[string]any
        if json.Unmarshal(b, &m) == nil && m["schema"] == "ami.sum/v1" {
            sum = m
        }
    }
    pkgs, _ := sum["packages"].(map[string]any)
    if pkgs == nil { pkgs = map[string]any{} }

    var updated []modUpdateItem
    // Copy each workspace package with a root, name, version
    for _, e := range ws.Packages {
        p := e.Package
        if p.Name == "" || p.Version == "" || p.Root == "" { continue }
        src := filepath.Clean(filepath.Join(dir, p.Root))
        if st, err := os.Stat(src); err != nil || !st.IsDir() { continue }
        dst := filepath.Join(cache, p.Name, p.Version)
        // Remove and copy
        if err := os.RemoveAll(dst); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: fmt.Sprintf("remove failed: %v", err)}) }
            return exit.New(exit.IO, "remove: %v", err)
        }
        if err := copyDir(src, dst); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: fmt.Sprintf("copy failed: %v", err)}) }
            return exit.New(exit.IO, "copy: %v", err)
        }
        // Hash and update sum
        h, err := hashDir(dst)
        if err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "hash failed"}) }
            return exit.New(exit.IO, "hash failed: %v", err)
        }
        pkgs[p.Name] = map[string]any{"version": p.Version, "sha256": h}
        updated = append(updated, modUpdateItem{Name: p.Name, Version: p.Version, Path: dst})
    }

    // Persist ami.sum with stable ordering via MarshalIndent
    sum["packages"] = pkgs
    if b, err := json.MarshalIndent(sum, "", "  "); err == nil {
        if err := os.WriteFile(sumPath, b, 0o644); err != nil {
            if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "write ami.sum failed"}) }
            return exit.New(exit.IO, "write ami.sum: %v", err)
        }
    } else {
        if jsonOut { _ = json.NewEncoder(out).Encode(modUpdateResult{Message: "encode ami.sum failed"}) }
        return exit.New(exit.Internal, "encode ami.sum: %v", err)
    }

    // Sort updated for deterministic output
    sort.Slice(updated, func(i, j int) bool {
        if updated[i].Name == updated[j].Name {
            return updated[i].Version < updated[j].Version
        }
        return updated[i].Name < updated[j].Name
    })

    if jsonOut {
        return json.NewEncoder(out).Encode(modUpdateResult{Updated: updated, Message: "ok"})
    }
    for _, u := range updated {
        _, _ = fmt.Fprintf(out, "updated %s@%s -> %s\n", u.Name, u.Version, u.Path)
    }
    return nil
}

