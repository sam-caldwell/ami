package driver

import (
    "path/filepath"
    "strings"
)

// unitName returns the filename without its extension.
func unitName(path string) string {
    base := filepath.Base(path)
    if i := strings.LastIndexByte(base, '.'); i > 0 { return base[:i] }
    return base
}

