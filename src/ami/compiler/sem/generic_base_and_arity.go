package sem

import "strings"

// genericBaseAndArity returns the generic base name and top-level argument count
// for a textual type like "Owned<int>" or "map<string,int>". If the type is not
// a generic form, ok will be false.
func genericBaseAndArity(s string) (base string, arity int, ok bool) {
    s = strings.TrimSpace(s)
    i := strings.IndexByte(s, '<')
    j := strings.LastIndexByte(s, '>')
    if i < 0 || j <= i { return "", 0, false }
    base = strings.TrimSpace(s[:i])
    inner := s[i+1 : j]
    // count top-level commas in inner, respecting nested <>
    depth := 0
    arity = 1
    for k := 0; k < len(inner); k++ {
        c := inner[k]
        switch c {
        case '<':
            depth++
        case '>':
            if depth > 0 { depth-- }
        case ',':
            if depth == 0 { arity++ }
        }
    }
    // empty inner → special case: zero args
    if strings.TrimSpace(inner) == "" { arity = 0 }
    return base, arity, true
}

