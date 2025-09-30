package scheduler

import "strings"

// ParsePolicy converts a string to a Policy, case-insensitive.
func ParsePolicy(s string) (Policy, bool) {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "fifo": return FIFO, true
    case "lifo": return LIFO, true
    case "fair": return FAIR, true
    case "worksteal", "work_steal", "work-steal": return WORKSTEAL, true
    default:
        return "", false
    }
}

