package driver

import "strings"

// lookupLetAssign finds an assignment to the given identifier name within the body and returns its RHS.
// Supports: let name = RHS; var name = RHS; name := RHS; name = RHS
func lookupLetAssign(body, name string) (string, bool) {
    lines := strings.Split(body, "\n")
    for _, ln := range lines {
        line := strings.TrimSpace(ln)
        if line == "" { continue }
        // normalize prefixes
        if strings.HasPrefix(line, "let ") { line = strings.TrimSpace(line[len("let "):]) }
        if strings.HasPrefix(line, "var ") { line = strings.TrimSpace(line[len("var "):]) }
        if !strings.HasPrefix(line, name) { continue }
        rest := strings.TrimSpace(line[len(name):])
        if strings.HasPrefix(rest, ":=") { rest = strings.TrimSpace(rest[2:]) } else if strings.HasPrefix(rest, "=") { rest = strings.TrimSpace(rest[1:]) } else { continue }
        if j := strings.IndexByte(rest, ';'); j >= 0 { rest = strings.TrimSpace(rest[:j]) }
        if rest != "" { return rest, true }
    }
    return "", false
}

