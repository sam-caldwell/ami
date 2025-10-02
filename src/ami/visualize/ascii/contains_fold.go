package ascii

import "strings"

// containsFold returns true if a contains b (case-insensitive ASCII fold).
func containsFold(a, b string) bool { return strings.Contains(strings.ToLower(a), strings.ToLower(b)) }

