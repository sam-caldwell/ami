package main

import (
    "path"
    "regexp"
    "strings"
)

// matchAnyRule returns true if code matches any of patterns.
// Supports substring, glob (*?[) and regex (re:expr or /expr/)
func matchAnyRule(code string, patterns []string) bool {
    for _, p := range patterns {
        if p == "" { continue }
        if len(p) >= 3 && p[0] == '/' && p[len(p)-1] == '/' {
            re, err := regexp.Compile(p[1:len(p)-1]); if err == nil && re.MatchString(code) { return true }
            continue
        }
        if strings.HasPrefix(p, "re:") {
            re, err := regexp.Compile(p[3:]); if err == nil && re.MatchString(code) { return true }
            continue
        }
        if strings.ContainsAny(p, "*?[") {
            if ok, _ := path.Match(p, code); ok { return true }
            continue
        }
        if strings.Contains(code, p) { return true }
    }
    return false
}

