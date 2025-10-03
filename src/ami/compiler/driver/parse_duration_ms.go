package driver

import (
    "strconv"
    "strings"
)

func parseDurationMs(s string) (int, bool) {
    // normalize spacing before trimming optional quotes
    s = trimQuotes(strings.TrimSpace(s))
    if n, err := strconv.Atoi(s); err == nil { return n, true }
    mul := 1
    if strings.HasSuffix(s, "ms") { mul = 1; s = strings.TrimSuffix(s, "ms") } else
    if strings.HasSuffix(s, "s") { mul = 1000; s = strings.TrimSuffix(s, "s") } else
    if strings.HasSuffix(s, "m") { mul = 60*1000; s = strings.TrimSuffix(s, "m") } else
    if strings.HasSuffix(s, "h") { mul = 60*60*1000; s = strings.TrimSuffix(s, "h") }
    n, err := strconv.Atoi(s)
    if err != nil { return 0, false }
    return n * mul, true
}
