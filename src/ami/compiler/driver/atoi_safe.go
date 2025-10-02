package driver

import "strconv"

func atoiSafe(s string) (int, bool) {
    s = trimQuotes(s)
    n, err := strconv.Atoi(s)
    if err != nil { return 0, false }
    return n, true
}

