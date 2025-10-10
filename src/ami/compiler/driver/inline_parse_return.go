package driver

import "strings"

// parseInlineReturn scans for a single-line "return ..." and parses supported forms.
func parseInlineReturn(body string) (returnParse, bool) {
    var rp returnParse
    if body == "" { return rp, false }
    lines := strings.Split(body, "\n")
    var line string
    for _, ln := range lines {
        if strings.Contains(ln, "return") { line = ln; break }
    }
    if line == "" { return rp, false }
    i := strings.Index(line, "return")
    rest := strings.TrimSpace(line[i+len("return"):])
    rest = strings.TrimSuffix(rest, ";")
    if j := strings.LastIndex(rest, ","); j >= 0 {
        tail := strings.TrimSpace(rest[j+1:])
        if strings.EqualFold(tail, "nil") { rest = strings.TrimSpace(rest[:j]) }
    }
    if rest == "ev" { rp.kind = retEV; return rp, true }
    // event.payload.field.<path>
    if strings.HasPrefix(rest, "event.payload.field.") {
        rp.kind = retField
        rp.path = strings.TrimPrefix(rest, "event.payload.field.")
        return rp, true
    }
    toks := fieldsPreserveQuotes(rest)
    if len(toks) == 3 {
        op := toks[1]
        if op == "+" || op == "-" || op == "*" || op == "/" || op == "%" {
            lraw := strings.TrimSpace(toks[0])
            rraw := strings.TrimSpace(toks[2])
            lhs := stripQuotes(lraw)
            rhs := stripQuotes(rraw)
            if lraw == "ev" && isLiteral(rhs) {
                rp.kind, rp.lhs, rp.rhs, rp.op, rp.lhsIsEv = retBinOp, "", rhs, op, true
                return rp, true
            }
            if rraw == "ev" && isLiteral(lhs) {
                rp.kind, rp.lhs, rp.rhs, rp.op, rp.rhsIsEv = retBinOp, lhs, "", op, true
                return rp, true
            }
            if isLiteral(lhs) && isLiteral(rhs) {
                rp.kind, rp.lhs, rp.rhs, rp.op = retBinOp, lhs, rhs, op
                return rp, true
            }
        }
        if op == "==" || op == "!=" || op == "<" || op == "<=" || op == ">" || op == ">=" {
            lhs := stripQuotes(toks[0])
            rhs := stripQuotes(toks[2])
            if isLiteral(lhs) && isLiteral(rhs) {
                rp.kind, rp.lhs, rp.rhs, rp.op = retCmp, lhs, rhs, op
                return rp, true
            }
        }
    }
    // Unary negation / logical not
    if strings.HasPrefix(rest, "-") {
        val := strings.TrimSpace(rest[1:])
        if val == "ev" || isLiteral(val) {
            // encode as binop with 0 - val for reuse downstream
            rp.kind = retUnary
            rp.op = "neg"
            if val == "ev" { rp.lhsIsEv = true } else { rp.lhs = stripQuotes(val) }
            return rp, true
        }
    }
    if strings.HasPrefix(rest, "!") {
        val := strings.TrimSpace(rest[1:])
        if val == "ev" || isLiteral(val) {
            rp.kind = retUnary
            rp.op = "not"
            if val == "ev" { rp.lhsIsEv = true } else { rp.lhs = stripQuotes(val) }
            return rp, true
        }
    }
    if isLiteral(rest) { rp.kind, rp.lit = retLit, stripQuotes(rest); return rp, true }
    return rp, false
}
