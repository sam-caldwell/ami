package sem

import (
    "strconv"
    "strings"
)

func validNonNegativeInt(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" { return false }
    n, err := strconv.Atoi(s)
    if err != nil { return false }
    return n >= 0
}

