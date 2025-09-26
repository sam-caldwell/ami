package parser

import (
    "unicode"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
)

// parseWorkerRef tries to interpret an argument string as a worker/factory reference.
func parseWorkerRef(s string) (astpkg.WorkerRef, bool) {
    name := s
    hasCall := false
    if i := indexRune(s, '('); i >= 0 {
        name = trimSpace(s[:i])
        hasCall = true
    }
    if !isIdentLexeme(name) { return astpkg.WorkerRef{}, false }
    kind := "function"
    if hasCall || hasPrefix(name, "New") { kind = "factory" }
    return astpkg.WorkerRef{Name: name, Kind: kind}, true
}

func isIdentLexeme(s string) bool {
    if s == "" { return false }
    r0 := []rune(s)[0]
    if !(unicode.IsLetter(r0) || r0 == '_') { return false }
    for _, r := range s[1:] {
        if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') { return false }
    }
    return true
}

// string helpers wrapping stdlib to avoid extra imports in this file
func indexRune(s string, r rune) int { for i, rr := range s { if rr == r { return i } }; return -1 }
func trimSpace(s string) string { i, j := 0, len(s); for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') { i++ }; for i < j && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') { j-- }; return s[i:j] }
func hasPrefix(s, pref string) bool { if len(pref) > len(s) { return false }; return s[:len(pref)] == pref }

