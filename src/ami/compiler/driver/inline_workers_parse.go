package driver

import (
    "strings"
)

// Helpers to extract and parse a tiny subset of inline func literal bodies.
// Supported now:
//   return ev
//   return <lit>
//   return <lit> <op> <lit>
// where lit is int, real, bool, or string (quoted). Optional trailing ", nil" is ignored.

// extractInlineBody returns the string inside the outermost braces of a func literal.
func extractInlineBody(s string) string {
    i := strings.Index(s, "{")
    if i < 0 { return "" }
    depth := 0
    for idx := i; idx < len(s); idx++ {
        if s[idx] == '{' { depth++ }
        if s[idx] == '}' {
            depth--
            if depth == 0 {
                inner := s[i+1 : idx]
                return strings.TrimSpace(inner)
            }
        }
    }
    return ""
}

type retKind int
const (
    retNone retKind = iota
    retEV
    retLit
    retBinOp
    retCmp
)

type returnParse struct {
    kind retKind
    lit  string
    lhs  string
    rhs  string
    op   string // one of +,-,*,/,%
    lhsIsEv bool
    rhsIsEv bool
}

// parseInlineReturn scans for a single-line "return ..." and parses supported forms.
func parseInlineReturn(body string) (returnParse, bool) {
    var rp returnParse
    if body == "" { return rp, false }
    // take first line containing return
    lines := strings.Split(body, "\n")
    var line string
    for _, ln := range lines {
        if strings.Contains(ln, "return") { line = ln; break }
    }
    if line == "" { return rp, false }
    i := strings.Index(line, "return")
    rest := strings.TrimSpace(line[i+len("return"):])
    // Strip optional trailing ;
    rest = strings.TrimSuffix(rest, ";")
    // Strip trailing ", nil" or ",nil"
    if j := strings.LastIndex(rest, ","); j >= 0 {
        tail := strings.TrimSpace(rest[j+1:])
        if strings.EqualFold(tail, "nil") { rest = strings.TrimSpace(rest[:j]) }
    }
    // Identity: ev
    if rest == "ev" { rp.kind = retEV; return rp, true }
    // Binary op: <lit> <op> <lit>
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
        // Comparisons: ==, !=, <, <=, >, >= on numeric literals
        if op == "==" || op == "!=" || op == "<" || op == "<=" || op == ">" || op == ">=" {
            lhs := stripQuotes(toks[0])
            rhs := stripQuotes(toks[2])
            if isLiteral(lhs) && isLiteral(rhs) {
                rp.kind, rp.lhs, rp.rhs, rp.op = retCmp, lhs, rhs, op
                return rp, true
            }
        }
    }
    // Single literal
    if isLiteral(rest) { rp.kind, rp.lit = retLit, stripQuotes(rest); return rp, true }
    return rp, false
}

func fieldsPreserveQuotes(s string) []string {
    var out []string
    var cur strings.Builder
    inStr := false
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch == '"' { inStr = !inStr; cur.WriteByte(ch); continue }
        if !inStr && (ch == ' ' || ch == '\t') {
            if cur.Len() > 0 { out = append(out, strings.TrimSpace(cur.String())); cur.Reset() }
            continue
        }
        cur.WriteByte(ch)
    }
    if cur.Len() > 0 { out = append(out, strings.TrimSpace(cur.String())) }
    return out
}

func isLiteral(s string) bool {
    s = strings.TrimSpace(s)
    if s == "" { return false }
    if s == "true" || s == "false" { return true }
    if s[0] == '"' && s[len(s)-1] == '"' { return true }
    // crude numeric check: digits and optional dot
    dot := false
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch >= '0' && ch <= '9' { continue }
        if ch == '.' && !dot { dot = true; continue }
        if i == 0 && (ch == '+' || ch == '-') { continue }
        return false
    }
    return true
}

func stripQuotes(s string) string {
    s = strings.TrimSpace(s)
    if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' { return s[1:len(s)-1] }
    return s
}

func isPrimitiveLike(t string) bool {
    tt := strings.TrimSpace(t)
    switch tt {
    case "bool", "int", "int64", "uint64", "real", "float64", "string":
        return true
    default:
        return false
    }
}

func isNumericLike(t string) bool {
    tt := strings.TrimSpace(t)
    switch tt {
    case "int", "int64", "uint64", "real", "float64":
        return true
    default:
        return false
    }
}
