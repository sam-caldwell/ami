package enum

// IsValid reports whether v is within the enum range.
func IsValid(d Descriptor, v int) bool { return v >= 0 && v < len(d.Names) }

