package main

import "strings"

// splitImportConstraint splits an import entry like "path >= v1.2.3" into
// path and constraint parts. When no constraint is recognized, returns the
// original entry and empty constraint.
func splitImportConstraint(entry string) (string, string) {
    parts := strings.Fields(entry)
    if len(parts) <= 1 { return entry, "" }
    path := parts[0]
    constraint := strings.TrimSpace(strings.Join(parts[1:], " "))
    return path, constraint
}

