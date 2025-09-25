package filepath

import stdstrings "strings"

// normalizeSeparators converts all Windows-style path separators to '/'
// to ensure cross-platform deterministic behavior.
func normalizeSeparators(s string) string { return stdstrings.ReplaceAll(s, "\\", "/") }

