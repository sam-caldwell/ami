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

