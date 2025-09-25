package mod

import (
    "path/filepath"
    "strings"
)

// Backend defines a pluggable registry backend capable of fetching
// modules into the local cache.
type Backend interface {
    // Name returns a short identifier for the backend (e.g., "file", "git+ssh").
    Name() string
    // Match reports whether the backend can handle the provided spec/URL.
    Match(spec string) bool
    // Fetch downloads or stages the module into the provided cache directory,
    // returning the destination path, and optionally the package and version.
    // If the backend does not participate in ami.sum updates, pkg and ver
    // should be empty strings.
    Fetch(spec, cacheDir string) (dest string, pkg string, ver string, err error)
}

var backends []Backend

// registerBackend adds a backend to the registry. Called from init() of each backend implementation.
func registerBackend(b Backend) { backends = append(backends, b) }

// selectBackend chooses the first backend whose Match returns true for the spec.
func selectBackend(spec string) Backend {
    // Normalize simple local paths to use forward slashes, but do not modify URL-like specs.
    isLocal := strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") || strings.HasPrefix(spec, "/") || strings.HasPrefix(spec, "file://")
    normalized := spec
    if isLocal && !strings.HasPrefix(spec, "file://") {
        normalized = filepath.ToSlash(spec)
    }
    for _, b := range backends {
        if b.Match(normalized) { return b }
    }
    return nil
}

