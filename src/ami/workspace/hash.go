package workspace

import (
    "crypto/sha256"
    "encoding/hex"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
)

// HashDir returns a deterministic sha256 across file contents in a directory.
// It sorts file paths, and for each file appends the relative path then bytes.
func HashDir(root string) (string, error) {
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
        b, err := os.ReadFile(filepath.Join(root, rel))
        if err != nil { return "", err }
        _, _ = h.Write([]byte(rel))
        _, _ = h.Write(b)
    }
    return hex.EncodeToString(h.Sum(nil)), nil
}

