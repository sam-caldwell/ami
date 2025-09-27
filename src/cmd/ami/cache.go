package main

import (
    "os"
    "path/filepath"

    "github.com/sam-caldwell/ami/src/ami/exit"
)

// ensurePackageCache detects the package cache directory and ensures it exists.
// Behavior:
// - If AMI_PACKAGE_CACHE is set, create it if missing.
// - If not set, default to ${HOME}/.ami/pkg and create it if missing.
func ensurePackageCache() error {
    cache := os.Getenv("AMI_PACKAGE_CACHE")
    if cache == "" {
        home, err := os.UserHomeDir()
        if err != nil {
            // HOME not available (e.g., some CI/test environments). Do not fail root command.
            return nil
        }
        cache = filepath.Join(home, ".ami", "pkg")
        // Persist the derived path for child processes (tests may rely on it)
        _ = os.Setenv("AMI_PACKAGE_CACHE", cache)
    }
    if _, err := os.Stat(cache); os.IsNotExist(err) {
        if mkErr := os.MkdirAll(cache, 0o755); mkErr != nil {
            return exit.New(exit.IO, "create package cache: %v", mkErr)
        }
    }
    return nil
}
