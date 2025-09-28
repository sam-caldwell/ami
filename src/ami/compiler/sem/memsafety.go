package sem

import (
    "time"

    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// AnalyzeMemorySafety scans a file's tokens to enforce AMI 2.3.2 memory rules:
// - Ban '&' operator anywhere (no raw address-of)
// - Unary '*' is not dereference; only allowed as a mutating marker on LHS: "* ident = ..."
// It emits diag records with stable codes and positions.
func AnalyzeMemorySafety(f *source.File) []diag.Record {
    var out []diag.Record
    if f == nil { return out }
    s := scanner.New(f)
    // collect tokens to allow lookahead
    var toks []token.Token
    for {
        t := s.Next()
        toks = append(toks, t)
        if t.Kind == token.EOF { break }
    }
    now := time.Unix(0, 0).UTC()
    // helper to check if token represents an operand end (for unary detection)
    isOperand := func(k token.Kind) bool {
        switch k {
        case token.Ident, token.Number, token.String, token.RParenSym, token.RBracketSym, token.RBraceSym:
            return true
        default:
            return false
        }
    }
    nextNonComment := func(i int) int {
        j := i+1
        for j < len(toks) {
            if toks[j].Kind == token.LineComment || toks[j].Kind == token.BlockComment { j++; continue }
            return j
        }
        return len(toks)-1
    }
    prevNonComment := func(i int) int {
        j := i-1
        for j >= 0 {
            if toks[j].Kind == token.LineComment || toks[j].Kind == token.BlockComment { j--; continue }
            return j
        }
        return -1
    }
    for i := 0; i < len(toks); i++ {
        t := toks[i]
        switch t.Kind {
        case token.Symbol, token.BitAnd:
            // Ban unary address-of '&' but allow binary bitwise-and.
            if t.Kind == token.Symbol && t.Lexeme != "&" { break }
            pi := prevNonComment(i)
            unary := true
            if pi >= 0 {
                if toks[pi].Pos.Line == t.Pos.Line && isOperand(toks[pi].Kind) { unary = false }
            }
            if unary {
                out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "address-of operator '&' is not allowed in AMI", File: f.Name, Pos: &diag.Position{Line: t.Pos.Line, Column: t.Pos.Column, Offset: t.Pos.Offset}})
            }
        case token.Star:
            // Determine unary vs binary. Treat '*' at start of a new line as unary.
            pi := prevNonComment(i)
            unary := true
            if pi >= 0 {
                // If previous token is on the same line and is an operand, this is binary '*'
                if toks[pi].Pos.Line == t.Pos.Line && isOperand(toks[pi].Kind) {
                    unary = false
                }
            }
            if !unary { continue }
            // expect: * Ident =
            ni := nextNonComment(i)
            if ni < len(toks) && toks[ni].Kind == token.Ident {
                n2 := nextNonComment(ni)
                if n2 < len(toks) && toks[n2].Kind == token.Assign {
                    // allowed pattern
                    continue
                }
            }
            out = append(out, diag.Record{Timestamp: now, Level: diag.Error, Code: "E_MUT_BLOCK_UNSUPPORTED", Message: "unary '*' only allowed as LHS mutating marker '* ident = ...'", File: f.Name, Pos: &diag.Position{Line: t.Pos.Line, Column: t.Pos.Column, Offset: t.Pos.Offset}})
        }
    }
    return out
}
