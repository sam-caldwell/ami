package sem

import "strings"

// isDurationLike reports whether s looks like a simple duration (e.g., 100ms, 2s, 3m, 1h).
func isDurationLike(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" { return false }
    return durRe.MatchString(s)
}

