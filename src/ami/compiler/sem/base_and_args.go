package sem

import "strings"

func baseAndArgs(s string) (string, []string, bool) {
    s = strings.TrimSpace(s)
    i := strings.IndexByte(s, '<')
    if i < 0 || !strings.HasSuffix(s, ">") { return "", nil, false }
    base := strings.TrimSpace(s[:i])
    inner := s[i+1 : len(s)-1]
    if strings.TrimSpace(inner) == "" { return base, []string{}, true }
    parts := splitTop(inner)
    return base, parts, true
}

