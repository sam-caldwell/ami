package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseAttrArg parses a single attribute argument, accepting either
// - an expression (ident/selector/call/number/string/literal), or
// - a key=value pair where key is an identifier and value is an expression.
func (p *Parser) parseAttrArg() (ast.Arg, bool) {
    if p.cur.Kind == token.Ident || p.cur.Kind == token.KwType {
        key := p.cur.Lexeme
        pos := p.cur.Pos
        p.next()
        if p.cur.Kind == token.Assign {
            // k = value
            p.next()
            // Special-case bracketed list forms: [a,b,c]
            if p.cur.Kind == token.LBracketSym {
                start := p.cur.Pos.Offset
                depth := 0
                var last int
                for {
                    if p.cur.Kind == token.EOF { break }
                    if p.cur.Kind == token.LBracketSym { depth++ }
                    if p.cur.Kind == token.RBracketSym {
                        if depth > 0 { depth-- }
                        last = p.cur.Pos.Offset
                        p.next()
                        if depth == 0 { break }
                        continue
                    }
                    p.next()
                }
                src := p.s.FileContent()
                val := "[]"
                if last > start && last+1 <= len(src) {
                    val = src[start : last+1]
                }
                return ast.Arg{Pos: pos, Text: key + "=" + val}, true
            }
            // Special-case function literal: func(...) [results] { ... }
            if p.cur.Kind == token.KwFunc {
                start := p.cur.Pos.Offset
                // consume 'func'
                p.next()
                // params (balanced parentheses)
                if p.cur.Kind == token.LParenSym {
                    depth := 0
                    for {
                        if p.cur.Kind == token.EOF { break }
                        if p.cur.Kind == token.LParenSym { depth++ }
                        if p.cur.Kind == token.RParenSym {
                            if depth > 0 { depth-- }
                            p.next()
                            if depth == 0 { break }
                            continue
                        }
                        p.next()
                    }
                }
                // optional results: either parenthesized or single type
                if p.cur.Kind == token.LParenSym {
                    depth := 0
                    for {
                        if p.cur.Kind == token.EOF { break }
                        if p.cur.Kind == token.LParenSym { depth++ }
                        if p.cur.Kind == token.RParenSym {
                            if depth > 0 { depth-- }
                            p.next()
                            if depth == 0 { break }
                            continue
                        }
                        p.next()
                    }
                } else if p.isTypeName(p.cur.Kind) || p.cur.Kind == token.Ident {
                    // simple single result type; allow qualified like pkg.Type and generic angles
                    p.next()
                    for p.cur.Kind == token.DotSym { p.next(); if p.cur.Kind == token.Ident { p.next() } else { break } }
                    if p.cur.Kind == token.Lt || p.cur.Kind == token.Shl {
                        adepth := 0
                        for {
                            if p.cur.Kind == token.EOF { break }
                            switch p.cur.Kind {
                            case token.Lt:
                                adepth++
                            case token.Shl:
                                adepth += 2
                            case token.Gt:
                                adepth--
                                p.next()
                                if adepth == 0 { break }
                                continue
                            case token.Shr:
                                adepth -= 2
                                p.next()
                                if adepth == 0 { break }
                                continue
                            }
                            p.next()
                        }
                    }
                }
                // body
                if p.cur.Kind == token.LBraceSym {
                    depth := 0
                    var last int
                    for {
                        if p.cur.Kind == token.EOF { break }
                        if p.cur.Kind == token.LBraceSym { depth++ }
                        if p.cur.Kind == token.RBraceSym {
                            if depth > 0 { depth-- }
                            last = p.cur.Pos.Offset
                            p.next()
                            if depth == 0 { break }
                            continue
                        }
                        p.next()
                    }
                    src := p.s.FileContent()
                    val := "func{}"
                    if last >= start && last+1 <= len(src) {
                        val = src[start : last+1]
                    }
                    return ast.Arg{Pos: pos, Text: key + "=" + val}, true
                }
                // if no body, still capture header till current position
                src := p.s.FileContent()
                cur := p.cur.Pos.Offset
                if cur <= start { cur = start }
                if cur > len(src) { cur = len(src) }
                val := src[start:cur]
                return ast.Arg{Pos: pos, Text: key + "=" + val}, true
            }
            // Permit bare primitive/alias type keywords and certain domain keywords (not plain identifiers)
            if p.isAttrBareValueKeyword(p.cur.Kind) {
                val := p.cur.Lexeme
                p.next()
                return ast.Arg{Pos: pos, Text: key + "=" + val}, true
            }
            e, ok := p.parseExprPrec(1)
            if !ok {
                p.errf("expected value expression after '=', got %q", p.cur.Lexeme)
                return ast.Arg{}, false
            }
            return ast.Arg{Pos: pos, Text: key + "=" + exprText(e)}, true
        }
        // otherwise treat as ident-led expression (selector/call allowed)
        e := p.parseIdentExpr(key, pos)
        return ast.Arg{Pos: ePos(e), Text: exprText(e), IsString: isStringLit(e)}, true
    }
    // fallback to generic expression
    e, ok := p.parseExprPrec(1)
    if !ok {
        return ast.Arg{}, false
    }
    return ast.Arg{Pos: ePos(e), Text: exprText(e), IsString: isStringLit(e)}, true
}
