package root

import "strings"

func ruleSelected(code string) bool {
    if strings.TrimSpace(lintRules) == "" {
        return true
    }
    // code is like E_X or W_Y; do case-insensitive substring match
    parts := strings.Split(lintRules, ",")
    lc := strings.ToLower(code)
    for _, p := range parts {
        if strings.Contains(lc, strings.ToLower(strings.TrimSpace(p))) {
            return true
        }
    }
    return false
}

