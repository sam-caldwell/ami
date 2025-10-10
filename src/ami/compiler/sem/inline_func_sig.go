package sem

import "strings"

// inlineFuncSig parses a function literal header like:
//   func(param Event<T>) (Event<U>, error) { ... }
//   func(ev Event<T>) error { ... }
// and returns the parameter type string (as written) and a slice of
// result type strings (1 for single, 2 for tuple). It ignores the body.
// Returns ok=false when it cannot recognize the shape.
func inlineFuncSig(text string) (paramType string, results []string, ok bool) {
    s := strings.TrimSpace(text)
    i := strings.Index(s, "func")
    if i < 0 { return "", nil, false }
    s = s[i+len("func"):]
    s = strings.TrimSpace(s)
    if !strings.HasPrefix(s, "(") { return "", nil, false }
    // scan params to matching ')'
    depth := 0
    end := -1
    for idx, r := range s {
        if r == '(' { depth++ }
        if r == ')' {
            depth--
            if depth == 0 { end = idx; break }
        }
    }
    if end <= 0 { return "", nil, false }
    params := s[1:end]
    tail := strings.TrimSpace(s[end+1:])
    // derive param type: take last token outside angle brackets
    paramType = deriveParamType(strings.TrimSpace(params))

    // parse results
    if tail == "" || tail[0] == '{' {
        // no result section
        return paramType, nil, true
    }
    // parenthesized result list
    if tail[0] == '(' {
        d := 0
        e := -1
        for i2, r := range tail {
            if r == '(' { d++ }
            if r == ')' { d--; if d == 0 { e = i2; break } }
        }
        if e <= 0 { return paramType, nil, false }
        inner := strings.TrimSpace(tail[1:e])
        // split by comma outside angle brackets
        results = splitTopLevel(inner)
        for i := range results { results[i] = strings.TrimSpace(results[i]) }
        return paramType, results, true
    }
    // single result until '{'
    if i3 := strings.IndexByte(tail, '{'); i3 >= 0 {
        one := strings.TrimSpace(tail[:i3])
        if one != "" { return paramType, []string{one}, true }
        return paramType, nil, true
    }
    one := strings.TrimSpace(tail)
    if one != "" { return paramType, []string{one}, true }
    return paramType, nil, true
}
