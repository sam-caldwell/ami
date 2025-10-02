package sem

import "strings"

func innerGeneric(s string) string {
    // return content between first '<' and last '>'
    i := strings.IndexByte(s, '<')
    j := strings.LastIndexByte(s, '>')
    if i < 0 || j <= i { return s }
    return s[i+1 : j]
}

