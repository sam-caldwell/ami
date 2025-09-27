package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/exit"
    "github.com/sam-caldwell/ami/src/ami/workspace"
)

type cleanResult struct {
    Path    string   `json:"path"`
    Removed bool     `json:"removed"`
    Created bool     `json:"created"`
    Messages []string `json:"messages"`
}

// runClean removes and recreates the build target directory.
func runClean(out io.Writer, dir string, jsonOut bool) error {
    res := cleanResult{}
    if lg := getRootLogger(); lg != nil {
        lg.Info("clean.start", map[string]any{"dir": dir, "json": jsonOut})
    }

    // Determine target from workspace if present; otherwise default to ./build
    wsPath := filepath.Join(dir, "ami.workspace")
    var target string
    var ws workspace.Workspace
    if _, err := os.Stat(wsPath); errors.Is(err, os.ErrNotExist) {
        target = "./build"
        res.Messages = append(res.Messages, "workspace not found; using default ./build")
    } else if err == nil {
        if err := ws.Load(wsPath); err != nil {
            if jsonOut {
                _ = json.NewEncoder(out).Encode(res)
            }
            return exit.New(exit.IO, "failed to load workspace: %v", err)
        }
        if ws.Toolchain.Compiler.Target == "" {
            target = "./build"
        } else {
            target = ws.Toolchain.Compiler.Target
        }
    } else {
        if jsonOut {
            _ = json.NewEncoder(out).Encode(res)
        }
        return exit.New(exit.IO, "stat workspace: %v", err)
    }

    // Enforce relative path within workspace (no absolute path)
    if filepath.IsAbs(target) {
        if jsonOut {
            res.Path = target
            _ = json.NewEncoder(out).Encode(res)
        }
        if lg := getRootLogger(); lg != nil {
            lg.Warn("clean.absolute_target", map[string]any{"target": target})
        }
        return exit.New(exit.User, "toolchain.compiler.target must be a relative path: %s", target)
    }

    absPath := filepath.Clean(filepath.Join(dir, target))
    res.Path = absPath

    // Remove existing directory or file at target path.
    if err := os.RemoveAll(absPath); err != nil {
        if jsonOut {
            _ = json.NewEncoder(out).Encode(res)
        }
        return exit.New(exit.IO, "failed to remove build directory: %v", err)
    }
    res.Removed = true
    if lg := getRootLogger(); lg != nil {
        lg.Info("clean.removed", map[string]any{"path": absPath})
    }

    // Recreate directory with default permissions.
    if err := os.MkdirAll(absPath, 0o755); err != nil {
        if jsonOut {
            _ = json.NewEncoder(out).Encode(res)
        }
        return exit.New(exit.IO, "failed to create build directory: %v", err)
    }
    res.Created = true
    if lg := getRootLogger(); lg != nil {
        lg.Info("clean.created", map[string]any{"path": absPath})
    }

    if jsonOut {
        return json.NewEncoder(out).Encode(res)
    }
    // Human output
    fmt.Fprintf(out, "cleaned: %s\n", res.Path)
    return nil
}
