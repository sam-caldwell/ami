package sem

import "strings"

func isStructText(s string) bool {
    s = strings.TrimSpace(s)
    return strings.HasPrefix(s, "Struct{") && strings.HasSuffix(s, "}")
}

