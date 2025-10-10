package driver

import "strings"

// extractEventArg returns the generic type argument T for "Event<T>" or empty string if none.
func extractEventArg(t string) string {
    const pfx = "Event<"
    if !strings.HasPrefix(t, pfx) { return "" }
    s := t[len(pfx):]
    if i := strings.IndexByte(s, '>'); i >= 0 { return s[:i] }
    return ""
}

