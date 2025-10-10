package driver

import "strings"

// parseInlineReturnWithVars parses a return statement and substitutes simple identifiers
// with their RHS from prior let/assign statements in the body (one level of indirection).
// Supports forms like: return a + b; where a,b were assigned literals or 'ev'.
func parseInlineReturnWithVars(body string) (returnParse, bool) {
    var zero returnParse
    if body == "" { return zero, false }
    lines := strings.Split(body, "\n")
    var line string
    for _, ln := range lines { if strings.Contains(ln, "return") { line = ln; break } }
    if line == "" { return zero, false }
    i := strings.Index(line, "return")
    rest := strings.TrimSpace(line[i+len("return"):])
    rest = strings.TrimSuffix(rest, ";")
    if j := strings.LastIndex(rest, ","); j >= 0 {
        if strings.EqualFold(strings.TrimSpace(rest[j+1:]), "nil") { rest = strings.TrimSpace(rest[:j]) }
    }
    // Shortcut: direct payload-field expression
    if strings.HasPrefix(rest, "event.payload.field.") {
        return parseInlineReturn("return " + rest)
    }
    toks := fieldsPreserveQuotes(rest)
    // binary op form
    if len(toks) == 3 {
        lraw := strings.TrimSpace(toks[0]); op := toks[1]; rraw := strings.TrimSpace(toks[2])
        if op == "+" || op == "-" || op == "*" || op == "/" || op == "%" || op == "==" || op == "!=" || op == "<" || op == "<=" || op == ">" || op == ">=" {
            // substitute identifiers from lets
            if isIdent(lraw) && lraw != "ev" && !isLiteral(lraw) {
                if repl, ok := lookupLetAssign(body, lraw); ok { lraw = repl }
            }
            if isIdent(rraw) && rraw != "ev" && !isLiteral(rraw) {
                if repl, ok := lookupLetAssign(body, rraw); ok { rraw = repl }
            }
            return parseInlineReturn("return " + lraw + " " + op + " " + rraw)
        }
    }
    // unary forms
    if strings.HasPrefix(rest, "-") || strings.HasPrefix(rest, "!") {
        u := rest[0:1]; arg := strings.TrimSpace(rest[1:])
        if isIdent(arg) && arg != "ev" && !isLiteral(arg) {
            if repl, ok := lookupLetAssign(body, arg); ok { arg = repl }
        }
        return parseInlineReturn("return " + u + arg)
    }
    // single identifier
    if isIdent(rest) {
        if repl, ok := lookupLetAssign(body, rest); ok {
            return parseInlineReturn("return " + repl)
        }
    }
    return zero, false
}

