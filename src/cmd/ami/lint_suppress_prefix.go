package main

import "path/filepath"

// hasPathPrefix reports whether path starts with prefix path segment-wise.
func hasPathPrefix(path, prefix string) bool {
    rel, err := filepath.Rel(prefix, path)
    if err != nil { return false }
    return rel == "." || (len(rel) > 0 && rel[0] != '.')
}

