package workspace

import (
    "os"
    "path/filepath"
)

// DefaultCacheRoot resolves the package cache root using AMI_PACKAGE_CACHE or default HOME path.
func DefaultCacheRoot() (string, error) {
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache != "" { return cache, nil }
    home, err := os.UserHomeDir()
    if err != nil { return "", err }
    return filepath.Join(home, ".ami", "pkg"), nil
}

