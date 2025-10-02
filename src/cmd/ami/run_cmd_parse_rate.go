package main

import (
    "strconv"
    "strings"
    "time"
)

func parseRate(rate string) time.Duration {
    if rate == "" { return 100 * time.Millisecond }
    if strings.Contains(rate, "/s") {
        nstr := strings.TrimSuffix(rate, "/s")
        if n, err := strconv.Atoi(nstr); err == nil && n > 0 { return time.Second / time.Duration(n) }
    }
    if d, err := time.ParseDuration(rate); err == nil { return d }
    return 100 * time.Millisecond
}

