package sem

import "strings"

// eventTypeArg extracts the inner type parameter from an Event<...> type string.
// Returns empty string if not in the expected form.
func eventTypeArg(typ string) string {
    const pfx = "Event<"
    if !strings.HasPrefix(typ, pfx) { return "" }
    s := typ[len(pfx):]
    if i := strings.IndexByte(s, '>'); i >= 0 { return s[:i] }
    return ""
}

