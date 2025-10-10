package driver

import "strings"

func stripQuotes(s string) string {
    s = strings.TrimSpace(s)
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' { return s[1:len(s)-1] }
    return s
}

