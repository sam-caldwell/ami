package scheduler

// Policy enumerates worker scheduling strategies.
type Policy string

const (
    FIFO      Policy = "fifo"
    LIFO      Policy = "lifo"
    FAIR      Policy = "fair"
    WORKSTEAL Policy = "worksteal"
)

// ParsePolicy normalizes a string to a Policy value.
func ParsePolicy(s string) (Policy, bool) {
    switch s {
    case "fifo", "FIFO":
        return FIFO, true
    case "lifo", "LIFO":
        return LIFO, true
    case "fair", "FAIR":
        return FAIR, true
    case "worksteal", "work-steal", "WORKSTEAL", "WORK-STEAL":
        return WORKSTEAL, true
    default:
        return "", false
    }
}
