package sem

import "strings"

// deriveParamType returns the parameter type token, tolerating an optional
// leading name (e.g., "ev Event<T>").
func deriveParamType(params string) string {
    // single parameter expected; conservatively take last token outside '<>'
    lastSep := -1
    depth := 0
    for i, r := range params {
        switch r {
        case '<': depth++
        case '>': if depth > 0 { depth-- }
        case ' ', '\t':
            if depth == 0 { lastSep = i }
        }
    }
    if lastSep >= 0 {
        return strings.TrimSpace(params[lastSep+1:])
    }
    return strings.TrimSpace(params)
}

