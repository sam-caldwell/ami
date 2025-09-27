package workspace

import "strings"

// NormalizeImports trims whitespace and removes duplicates from p.Import preserving order.
func NormalizeImports(p *Package) {
    if len(p.Import) == 0 { return }
    seen := make(map[string]struct{}, len(p.Import))
    out := make([]string, 0, len(p.Import))
    for _, s := range p.Import {
        t := strings.TrimSpace(s)
        if t == "" { continue }
        if _, ok := seen[t]; ok { continue }
        seen[t] = struct{}{}
        out = append(out, t)
    }
    p.Import = out
}

// ParseImportEntry splits an import entry into a path and optional constraint string.
// Accepts forms:
//  - "module"
//  - "module <constraint>"
//  - "./local/path"
// Constraint is returned as-is (unparsed) so callers can decide how to handle it.
func ParseImportEntry(entry string) (path string, constraint string) {
    s := strings.TrimSpace(entry)
    if s == "" { return "", "" }
    // Split on first whitespace to allow operators like ">= 1.2.3"
    i := strings.IndexFunc(s, func(r rune) bool { return r == ' ' || r == '\t' })
    if i < 0 { return s, "" }
    path = strings.TrimSpace(s[:i])
    constraint = strings.TrimSpace(s[i:])
    return path, constraint
}

