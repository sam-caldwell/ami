package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/workspace"
)

// initResult moved to init_result.go

// runInit performs the initialization in the given directory.
func runInit(out io.Writer, dir string, force bool, jsonOut bool) error {
    res := initResult{WorkspacePath: filepath.Join(dir, "ami.workspace")}
    if lg := getRootLogger(); lg != nil {
        lg.Info("init.start", map[string]any{"dir": dir, "force": force, "json": jsonOut})
    }

    // Load existing or create default workspace.
    var ws workspace.Workspace
    workspaceExisted := true
    if _, err := os.Stat(res.WorkspacePath); errors.Is(err, os.ErrNotExist) {
        workspaceExisted = false
        // Create a new workspace file.
        ws = workspace.DefaultWorkspace()
        if err := ws.Save(res.WorkspacePath); err != nil {
            return err
        }
        res.Created = true
        res.Messages = append(res.Messages, "created ami.workspace")
        if lg := getRootLogger(); lg != nil {
            lg.Info("init.created_workspace", map[string]any{"path": res.WorkspacePath})
        }
    } else if err == nil {
        // Exists
        if err := ws.Load(res.WorkspacePath); err != nil {
            return err
        }
        // If force, fill missing fields and write back only if changed.
        if force {
            before := ws
            changed := false
            // Ensure mandatory fields
            if ws.Version == "" {
                ws.Version = "1.0.0"
                changed = true
            }
            // Toolchain fields
            if ws.Toolchain.Compiler.Target == "" {
                ws.Toolchain.Compiler.Target = "./build"
                changed = true
            }
            if len(ws.Toolchain.Compiler.Env) == 0 {
                ws.Toolchain.Compiler.Env = workspace.DefaultWorkspace().Toolchain.Compiler.Env
                changed = true
            }
            if len(ws.Packages) == 0 || ws.FindPackage("main") == nil {
                // Add default main package if missing.
                def := workspace.DefaultWorkspace()
                // Append without overwriting existing entries.
                ws.Packages = append(ws.Packages, def.Packages[0])
                changed = true
            }
            if changed {
                if err := ws.Save(res.WorkspacePath); err != nil {
                    return err
                }
                // avoid unused var, capture to indicate an update happened
                _ = before
                res.Updated = true
                res.Messages = append(res.Messages, "updated missing fields in ami.workspace")
                if lg := getRootLogger(); lg != nil {
                    lg.Info("init.updated_workspace", map[string]any{"path": res.WorkspacePath})
                }
            }
        }
    } else {
        return err
    }

    // Ensure target directory exists.
    var targetCreated bool
    if ws.Toolchain.Compiler.Target == "" {
        ws.Toolchain.Compiler.Target = "./build"
    }
    if err := os.MkdirAll(filepath.Join(dir, ws.Toolchain.Compiler.Target), 0o755); err != nil {
        return err
    }
    // Best-effort: detect if it was created by checking existence before/after not tracked; treat as created when Created.
    targetCreated = true
    res.TargetDirCreated = targetCreated
    if lg := getRootLogger(); lg != nil {
        lg.Info("init.target_ready", map[string]any{"target": ws.Toolchain.Compiler.Target})
    }

    // Ensure package root directory exists for main package.
    var pkgDirCreated bool
    if p := ws.FindPackage("main"); p != nil {
        if p.Root == "" {
            p.Root = "./src"
        }
        if err := os.MkdirAll(filepath.Join(dir, p.Root), 0o755); err != nil {
            return err
        }
        pkgDirCreated = true
        if lg := getRootLogger(); lg != nil {
            lg.Info("init.pkgroot_ready", map[string]any{"root": p.Root})
        }
    }
    res.PackageDirCreated = pkgDirCreated

    // Git repo check
    if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
        res.GitStatus = "present"
    } else if force {
        // Attempt to initialize git if available.
        if _, err := exec.LookPath("git"); err == nil {
            cmd := exec.Command("git", "init")
            cmd.Dir = dir
            if err := cmd.Run(); err != nil {
                return fmt.Errorf("git init failed: %w", err)
            }
            res.GitStatus = "initialized"
            res.Messages = append(res.Messages, "initialized git repository")
        } else {
            res.GitStatus = "required"
            res.Messages = append(res.Messages, "git not found; repository not initialized")
        }
    } else {
        res.GitStatus = "required"
        // Enforce requirement per SPEC: print error when not a repo (non‑destructive).
        if jsonOut {
            enc := json.NewEncoder(out)
            _ = enc.Encode(res)
        }
        return fmt.Errorf("not a git repository; use --force to initialize")
    }

    // Ensure .gitignore has ./build
    giPath := filepath.Join(dir, ".gitignore")
    appendLineIfMissing(giPath, "./build\n")
    if lg := getRootLogger(); lg != nil {
        lg.Info("init.complete", map[string]any{"git": res.GitStatus})
    }

    if jsonOut {
        enc := json.NewEncoder(out)
        return enc.Encode(res)
    }
    // Human summary output
    fmt.Fprintf(out, "workspace: %s\n", res.WorkspacePath)
    if res.Created || !workspaceExisted {
        fmt.Fprintln(out, "created workspace file")
    } else if res.Updated {
        fmt.Fprintln(out, "updated workspace file (missing fields)")
    } else {
        fmt.Fprintln(out, "workspace file present")
    }
    fmt.Fprintf(out, "git: %s\n", res.GitStatus)
    return nil
}

// helper functions in separate files to preserve one-declaration-per-file pattern.
