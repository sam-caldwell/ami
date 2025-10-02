package merge

import "strings"

// extractPath reads dotted path from JSON-like payloads represented as map[string]any.
func extractPath(root any, path string) (any, bool) {
    if path == "" { return root, true }
    m, ok := root.(map[string]any)
    if !ok { return nil, false }
    cur := any(m)
    for _, seg := range strings.Split(path, ".") {
        mm, ok := cur.(map[string]any)
        if !ok { return nil, false }
        v, ok := mm[seg]
        if !ok { return nil, false }
        cur = v
    }
    return cur, true
}

