package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

type modCleanResult struct {
    Path    string `json:"path"`
    Removed bool   `json:"removed"`
    Created bool   `json:"created"`
}

func runModClean(out io.Writer, jsonOut bool) error {
    p := os.Getenv("AMI_PACKAGE_CACHE")
    if p == "" {
        home, err := os.UserHomeDir()
        if err != nil {
            // Fall back to a temp-based cache when HOME is unavailable (e.g., CI/test envs)
            p = filepath.Join(os.TempDir(), "ami", "pkg")
        } else {
            p = filepath.Join(home, ".ami", "pkg")
        }
    }
    abs := filepath.Clean(p)
    // Remove then recreate
    if err := os.RemoveAll(abs); err != nil {
        if jsonOut {
            _ = json.NewEncoder(out).Encode(modCleanResult{Path: abs})
        }
        return exit.New(exit.IO, "remove failed: %v", err)
    }
    if err := os.MkdirAll(abs, 0o755); err != nil {
        if jsonOut {
            _ = json.NewEncoder(out).Encode(modCleanResult{Path: abs, Removed: true})
        }
        return exit.New(exit.IO, "mkdir failed: %v", err)
    }
    res := modCleanResult{Path: abs, Removed: true, Created: true}
    if jsonOut {
        return json.NewEncoder(out).Encode(res)
    }
    _, _ = fmt.Fprintf(out, "cleaned: %s\n", abs)
    return nil
}
