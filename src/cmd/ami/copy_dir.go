package main

import (
    "io/fs"
    "os"
    "path/filepath"
)

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

