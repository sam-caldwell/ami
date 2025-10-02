package main

import (
    "crypto/sha256"
    "encoding/hex"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
)

func hashDir(root string) (string, error) {
    h := sha256.New()
    var files []string
    err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil { return err }
        if d.IsDir() { return nil }
        rel, err := filepath.Rel(root, path)
        if err != nil { return err }
        files = append(files, rel)
        return nil
    })
    if err != nil { return "", err }
    sort.Strings(files)
    for _, rel := range files {
        p := filepath.Join(root, rel)
        b, err := os.ReadFile(p)
        if err != nil { return "", err }
        _, _ = h.Write([]byte(rel))
        _, _ = h.Write(b)
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}

