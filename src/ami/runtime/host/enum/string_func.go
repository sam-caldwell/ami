package enum

// String returns the canonical name for ordinal v, or "" if invalid.
func String(d Descriptor, v int) string {
    if v >= 0 && v < len(d.Names) { return d.Names[v] }
    return ""
}

