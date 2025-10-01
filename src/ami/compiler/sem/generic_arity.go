package sem

import (
    "sort"
    "strings"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

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

// isGenericArityMismatch compares two textual types and reports whether they
// refer to the same generic base but with a different number of type arguments.
func isGenericArityMismatch(expected, actual string) (bool, string, int, int) {
    eb, ea, eok := genericBaseAndArity(expected)
    ab, aa, aok := genericBaseAndArity(actual)
    if !eok || !aok { return false, "", 0, 0 }
    if eb != ab { return false, "", 0, 0 }
    if ea != aa { return true, eb, ea, aa }
    return false, eb, ea, aa
}

// findGenericArityMismatchDeep parses expected and actual types and recursively
// searches for a generic arity mismatch at any nesting level. Returns the first
// mismatch encountered (base name and arities). Falls back to top‑level textual
// detection when parsing fails.
func findGenericArityMismatchDeep(expected, actual string) (bool, string, int, int) {
    // Text-based recursive scan that respects nested '<>' and splits top-level commas.
    eb, eargs, eok := baseAndArgs(expected)
    ab, aargs, aok := baseAndArgs(actual)
    if !eok || !aok {
        return isGenericArityMismatch(expected, actual)
    }
    if eb != ab { return false, "", 0, 0 }
    if len(eargs) != len(aargs) { return true, eb, len(eargs), len(aargs) }
    // Recurse into paired arguments
    for i := range eargs {
        if m, b, w, g := findGenericArityMismatchDeep(eargs[i], aargs[i]); m { return m, b, w, g }
    }
    return false, "", 0, 0
}

// findGenericArityMismatchDeepPath is like findGenericArityMismatchDeep but also returns
// a path of generic base names from the outermost to the mismatching base, and the
// argument indices taken at each nesting level.
func findGenericArityMismatchDeepPath(expected, actual string) (bool, []string, []int, string, int, int) {
    eb, eargs, eok := baseAndArgs(expected)
    ab, aargs, aok := baseAndArgs(actual)
    if !eok || !aok {
        // attempt typed detection (handles Struct)
        if m, path, pathIdx, _, b, w, g := findGenericArityMismatchWithFields(expected, actual); m {
            return true, path, pathIdx, b, w, g
        }
        if m, b, w, g := isGenericArityMismatch(expected, actual); m { return true, nil, nil, b, w, g }
        return false, nil, nil, "", 0, 0
    }
    if eb != ab { return false, nil, nil, "", 0, 0 }
    if len(eargs) != len(aargs) { return true, []string{eb}, []int{}, eb, len(eargs), len(aargs) }
    for i := range eargs {
        if m, p, idx, b, w, g := findGenericArityMismatchDeepPath(eargs[i], aargs[i]); m {
            // prepend current base
            path := append([]string{eb}, p...)
            pathIdx := append([]int{i}, idx...)
            return true, path, pathIdx, b, w, g
        }
    }
    return false, nil, nil, "", 0, 0
}

// findGenericArityMismatchWithFields parses both sides and finds a nested generic arity mismatch.
// Returns a path of generic bases, their argument indices per level, and a struct field path when applicable.
func findGenericArityMismatchWithFields(expected, actual string) (bool, []string, []int, []string, string, int, int) {
    et, eerr := types.Parse(expected)
    at, aerr := types.Parse(actual)
    if eerr != nil || aerr != nil { return false, nil, nil, nil, "", 0, 0 }
    return arityMismatchInTypesWithFields(et, at)
}

func arityMismatchInTypesWithFields(et, at types.Type) (bool, []string, []int, []string, string, int, int) {
    switch ev := et.(type) {
    case types.Generic:
        av, ok := at.(types.Generic); if !ok { return false, nil, nil, nil, "", 0, 0 }
        if ev.Name != av.Name { return false, nil, nil, nil, "", 0, 0 }
        if len(ev.Args) != len(av.Args) { return true, []string{ev.Name}, []int{}, nil, ev.Name, len(ev.Args), len(av.Args) }
        for i := range ev.Args {
            if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ev.Args[i], av.Args[i]); m {
                return true, append([]string{ev.Name}, p...), append([]int{i}, idx...), fp, b, w, g
            }
        }
        return false, nil, nil, nil, "", 0, 0
    case types.Optional:
        av, ok := at.(types.Optional); if !ok { return false, nil, nil, nil, "", 0, 0 }
        return arityMismatchInTypesWithFields(ev.Inner, av.Inner)
    case types.Struct:
        av, ok := at.(types.Struct); if !ok { return false, nil, nil, nil, "", 0, 0 }
        // check common fields deterministically
        keys := make([]string, 0, len(ev.Fields))
        for k := range ev.Fields { if _, ok := av.Fields[k]; ok { keys = append(keys, k) } }
        sort.Strings(keys)
        for _, k := range keys {
            if m, p, idx, fp, b, w, g := arityMismatchInTypesWithFields(ev.Fields[k], av.Fields[k]); m {
                // prepend struct marker and field name
                return true, p, idx, append([]string{"Struct", k}, fp...), b, w, g
            }
        }
        return false, nil, nil, nil, "", 0, 0
    case types.Named:
        name := ev.Name
        if name == "any" || (len(name) == 1 && name[0] >= 'A' && name[0] <= 'Z') { return false, nil, nil, nil, "", 0, 0 }
        return false, nil, nil, nil, "", 0, 0
    default:
        return false, nil, nil, nil, "", 0, 0
    }
}

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

func splitTop(s string) []string {
    var parts []string
    depth := 0
    inQuote := byte(0)
    last := 0
    for i := 0; i < len(s); i++ {
        c := s[i]
        if inQuote != 0 {
            if c == inQuote { inQuote = 0 }
            continue
        }
        switch c {
        case '\'', '"':
            inQuote = c
        case '<':
            depth++
        case '>':
            if depth > 0 { depth-- }
        case ',':
            if depth == 0 {
                parts = append(parts, strings.TrimSpace(s[last:i]))
                last = i + 1
            }
        }
    }
    tail := strings.TrimSpace(s[last:])
    if tail != "" { parts = append(parts, tail) }
    return parts
}

// findGenericArityMismatchDeepPathTextFields attempts to discover a generic arity
// mismatch and return both generic path/pathIdx and a textual struct fieldPath
// without relying on types.Parse. It handles simple Struct{...} and Optional<...>
// text forms and delegates generic path detection to findGenericArityMismatchDeepPath
// when not within a struct field.
func findGenericArityMismatchDeepPathTextFields(expected, actual string) (bool, []string, []int, []string, string, int, int) {
    // Optional wrappers
    if be, argsE, okE := baseAndArgs(expected); okE && be == "Optional" && len(argsE) == 1 {
        if ba, argsA, okA := baseAndArgs(actual); okA && ba == "Optional" && len(argsA) == 1 {
            return findGenericArityMismatchDeepPathTextFields(argsE[0], argsA[0])
        }
    }
    // Struct traversal
    if isStructText(expected) && isStructText(actual) {
        ef, okE := parseStructFieldsText(expected)
        af, okA := parseStructFieldsText(actual)
        if okE && okA {
            // common fields in stable order
            keys := make([]string, 0)
            for k := range ef { if _, ok := af[k]; ok { keys = append(keys, k) } }
            sort.Strings(keys)
            for _, k := range keys {
                einner := ef[k]
                ainner := af[k]
                if m, p, idx, fp, b, w, g := findGenericArityMismatchDeepPathTextFields(einner, ainner); m {
                    // prepend Struct → field
                    return true, p, idx, append([]string{"Struct", k}, fp...), b, w, g
                }
                if m2, p2, idx2, b2, w2, g2 := findGenericArityMismatchDeepPath(einner, ainner); m2 {
                    return true, p2, idx2, []string{"Struct", k}, b2, w2, g2
                }
            }
        }
    }
    // Fallback to generic deep path detection without fieldPath
    if m, p, idx, b, w, g := findGenericArityMismatchDeepPath(expected, actual); m {
        return true, p, idx, nil, b, w, g
    }
    return false, nil, nil, nil, "", 0, 0
}

func isStructText(s string) bool {
    s = strings.TrimSpace(s)
    return strings.HasPrefix(s, "Struct{") && strings.HasSuffix(s, "}")
}

func parseStructFieldsText(s string) (map[string]string, bool) {
    s = strings.TrimSpace(s)
    if !isStructText(s) { return nil, false }
    body := s[len("Struct{") : len(s)-1]
    out := map[string]string{}
    if strings.TrimSpace(body) == "" { return out, true }
    parts := splitTopAllText(body)
    for _, p := range parts {
        if i := strings.IndexByte(p, ':'); i > 0 {
            name := strings.TrimSpace(p[:i])
            ty := strings.TrimSpace(p[i+1:])
            if name != "" { out[name] = ty }
        } else {
            return nil, false
        }
    }
    return out, true
}

// splitTopAllText splits on commas at top level, respecting nested '<>', '{}', '()' and quotes.
func splitTopAllText(s string) []string {
    var parts []string
    depthAngle, depthBrace, depthParen := 0, 0, 0
    inQuote := byte(0)
    last := 0
    for i := 0; i < len(s); i++ {
        c := s[i]
        if inQuote != 0 {
            if c == inQuote { inQuote = 0 }
            continue
        }
        switch c {
        case '\'', '"':
            inQuote = c
        case '<':
            depthAngle++
        case '>':
            if depthAngle > 0 { depthAngle-- }
        case '{':
            depthBrace++
        case '}':
            if depthBrace > 0 { depthBrace-- }
        case '(':
            depthParen++
        case ')':
            if depthParen > 0 { depthParen-- }
        case ',':
            if depthAngle == 0 && depthBrace == 0 && depthParen == 0 {
                parts = append(parts, strings.TrimSpace(s[last:i]))
                last = i + 1
            }
        }
    }
    tail := strings.TrimSpace(s[last:])
    if tail != "" { parts = append(parts, tail) }
    return parts
}
