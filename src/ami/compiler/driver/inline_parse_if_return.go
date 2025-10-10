package driver

import "strings"

// parseInlineIfReturn recognizes: if <lit> <cmp> <lit> { return <lit|ev> [, nil] } else { return <lit|ev> [, nil] }
func parseInlineIfReturn(body string) (ifReturn, bool) {
    var out ifReturn
    s := strings.TrimSpace(body)
    if !strings.HasPrefix(s, "if ") { return out, false }
    s = s[len("if ") : ]
    i := strings.IndexByte(s, '{')
    if i < 0 { return out, false }
    cond := strings.TrimSpace(s[:i])
    rest := strings.TrimSpace(s[i:])
    toks := fieldsPreserveQuotes(cond)
    if len(toks) != 3 { return out, false }
    op := toks[1]
    if !(op == "==" || op == "!=" || op == "<" || op == "<=" || op == ">" || op == ">=") { return out, false }
    out.lhs, out.op, out.rhs = stripQuotes(toks[0]), op, stripQuotes(toks[2])
    depth := 0
    end := -1
    for idx := 0; idx < len(rest); idx++ {
        if rest[idx] == '{' { depth++ }
        if rest[idx] == '}' { depth--; if depth == 0 { end = idx; break } }
    }
    if end < 0 { return out, false }
    thenBlk := rest[1:end]
    tail := strings.TrimSpace(rest[end+1:])
    if !strings.HasPrefix(tail, "else") { return out, false }
    tail = strings.TrimSpace(tail[len("else"):])
    if !strings.HasPrefix(tail, "{") { return out, false }
    depth = 0
    eend := -1
    for idx := 0; idx < len(tail); idx++ {
        if tail[idx] == '{' { depth++ }
        if tail[idx] == '}' { depth--; if depth == 0 { eend = idx; break } }
    }
    if eend < 0 { return out, false }
    elseBlk := tail[1:eend]
    parseRet := func(b string) (isEv bool, lit string, ok bool) {
        b = strings.TrimSpace(b)
        j := strings.Index(b, "return")
        if j < 0 { return false, "", false }
        val := strings.TrimSpace(b[j+len("return"):])
        if k := strings.IndexByte(val, ';'); k >= 0 { val = strings.TrimSpace(val[:k]) }
        if idx := strings.LastIndex(val, ","); idx >= 0 {
            if strings.EqualFold(strings.TrimSpace(val[idx+1:]), "nil") { val = strings.TrimSpace(val[:idx]) }
        }
        if val == "ev" { return true, "", true }
        if isLiteral(val) { return false, stripQuotes(val), true }
        return false, "", false
    }
    if isEv, lit, ok := parseRet(thenBlk); ok { out.thenIsEv, out.thenLit = isEv, lit } else { return out, false }
    if isEv, lit, ok := parseRet(elseBlk); ok { out.elseIsEv, out.elseLit = isEv, lit } else { return out, false }
    return out, true
}

