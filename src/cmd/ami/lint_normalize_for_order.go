package main

import "strings"

func normalizeForOrder(ss []string) []string {
    out := make([]string, len(ss))
    for i, s := range ss {
        s = strings.TrimPrefix(s, "./")
        s = strings.TrimSuffix(s, "/")
        out[i] = strings.ToLower(s)
    }
    return out
}

