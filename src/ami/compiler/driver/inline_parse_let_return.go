package driver

import (
    "strings"
)

// parseInlineLetReturn recognizes tiny let/assignment patterns that culminate in returning the variable:
// Supported forms in the function literal body (single-variable):
//   let x = <expr>; return x
//   var x = <expr>; return x
//   x := <expr>; return x
//   x = <expr>; return x
// where <expr> matches the minimal subset already supported by parseInlineReturn (ev, literals,
// basic arithmetic, comparisons, unary -/!). Returns the parsed expression as a returnParse.
func parseInlineLetReturn(body string) (returnParse, bool) {
    var rp returnParse
    s := strings.TrimSpace(body)
    if s == "" { return rp, false }
    // find a 'return <name>' token
    var retVar string
    lines := strings.Split(s, "\n")
    for _, ln := range lines {
        if idx := strings.Index(ln, "return"); idx >= 0 {
            tail := strings.TrimSpace(ln[idx+len("return"):])
            // strip optional trailing ; and trailing , nil
            if j := strings.IndexByte(tail, ';'); j >= 0 { tail = strings.TrimSpace(tail[:j]) }
            if k := strings.LastIndex(tail, ","); k >= 0 {
                if strings.EqualFold(strings.TrimSpace(tail[k+1:]), "nil") { tail = strings.TrimSpace(tail[:k]) }
            }
            // variable name is a bare identifier
            if tail != "" && isIdent(tail) { retVar = tail; break }
        }
    }
    if retVar == "" { return rp, false }
    // find assignment to that variable earlier in the body
    // accepted: let x = RHS; var x = RHS; x := RHS; x = RHS
    var rhs string
    for _, ln := range lines {
        line := strings.TrimSpace(ln)
        if line == "" { continue }
        // normalize spaces around := and =
        if strings.HasPrefix(line, "let ") { line = strings.TrimSpace(line[len("let "):]) }
        if strings.HasPrefix(line, "var ") { line = strings.TrimSpace(line[len("var "):]) }
        if !strings.HasPrefix(line, retVar) { continue }
        rest := strings.TrimSpace(line[len(retVar):])
        if strings.HasPrefix(rest, ":=") { rest = strings.TrimSpace(rest[2:]) } else if strings.HasPrefix(rest, "=") { rest = strings.TrimSpace(rest[1:]) } else { continue }
        // cut at first ';'
        if j := strings.IndexByte(rest, ';'); j >= 0 { rest = strings.TrimSpace(rest[:j]) }
        if rest == "" { continue }
        rhs = rest
        break
    }
    if rhs == "" { return rp, false }
    // Resolve simple alias: if RHS is a bare identifier, try to resolve its assignment.
    if isIdent(rhs) && rhs != "ev" && !isLiteral(rhs) {
        if repl, ok := lookupLetAssign(body, rhs); ok { rhs = repl }
    }
    // Reuse parseInlineReturn by synthesizing a 'return <rhs>' line.
    if pr, ok := parseInlineReturn("return " + rhs); ok {
        return pr, true
    }
    return rp, false
}
