package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseExprPrec(minPrec int) (ast.Expr, bool) {
    switch p.cur.Kind {
    case token.LParenSym:
        // parenthesized expression for grouping
        p.next()
        inner, ok := p.parseExprPrec(1)
        if !ok {
            return nil, false
        }
        if p.cur.Kind != token.RParenSym {
            // tolerate missing ')'
        } else {
            p.next()
        }
        return p.parseWithTernary(inner, minPrec), true
    case token.Bang:
        // unary logical not
        pos := p.cur.Pos
        p.next()
        // parse operand with high precedence
        rhs, ok := p.parseExprPrec(6)
        if !ok {
            return nil, false
        }
        u := &ast.UnaryExpr{Pos: pos, Op: token.Bang, X: rhs}
        return p.parseWithTernary(u, minPrec), true
    case token.Minus:
        // unary negation
        pos := p.cur.Pos
        p.next()
        rhs, ok := p.parseExprPrec(6)
        if !ok {
            return nil, false
        }
        u := &ast.UnaryExpr{Pos: pos, Op: token.Minus, X: rhs}
        return p.parseWithTernary(u, minPrec), true
    case token.TildeSym:
        // unary bitwise not
        pos := p.cur.Pos
        p.next()
        rhs, ok := p.parseExprPrec(6)
        if !ok {
            return nil, false
        }
        u := &ast.UnaryExpr{Pos: pos, Op: token.TildeSym, X: rhs}
        return p.parseWithTernary(u, minPrec), true
    case token.Ident, token.KwSlice, token.KwSet, token.KwMap:
        name := p.cur.Lexeme
        npos := p.cur.Pos
        p.next()
        left := p.parseIdentExpr(name, npos)
        return p.parseWithTernary(left, minPrec), true
    case token.KwFunc:
        // Tolerate function literal expressions in attribute values: func(...) [results] { ... }
        pos := p.cur.Pos
        p.next()
        // params
        if p.cur.Kind == token.LParenSym {
            depth := 0
            for {
                if p.cur.Kind == token.EOF { break }
                if p.cur.Kind == token.LParenSym { depth++ }
                if p.cur.Kind == token.RParenSym {
                    if depth > 0 { depth--; p.next(); if depth == 0 { break }; continue }
                    break
                }
                p.next()
            }
            if p.cur.Kind == token.RParenSym { p.next() }
        }
        // single unparenthesized result or parenthesized list
        if p.cur.Kind == token.LParenSym {
            depth := 0
            for {
                if p.cur.Kind == token.EOF { break }
                if p.cur.Kind == token.LParenSym { depth++ }
                if p.cur.Kind == token.RParenSym {
                    if depth > 0 { depth--; p.next(); if depth == 0 { break }; continue }
                    break
                }
                p.next()
            }
            if p.cur.Kind == token.RParenSym { p.next() }
        } else if p.isTypeName(p.cur.Kind) || p.cur.Kind == token.Ident {
            // consume a simple result type, allow qualified/generic
            p.next()
            for p.cur.Kind == token.DotSym { p.next(); if p.cur.Kind == token.Ident { p.next() } else { break } }
            if p.cur.Kind == token.Lt || p.cur.Kind == token.Shl {
                depth := 0
                for {
                    if p.cur.Kind == token.EOF { break }
                    switch p.cur.Kind {
                    case token.Lt:
                        depth++
                    case token.Shl:
                        depth += 2
                    case token.Gt:
                        depth--
                        p.next()
                        if depth == 0 { break }
                        continue
                    case token.Shr:
                        depth -= 2
                        p.next()
                        if depth == 0 { break }
                        continue
                    }
                    p.next()
                }
            }
        }
        // body
        if p.cur.Kind == token.LBraceSym {
            depth := 0
            for {
                if p.cur.Kind == token.EOF { break }
                if p.cur.Kind == token.LBraceSym { depth++ }
                if p.cur.Kind == token.RBraceSym { depth--; p.next(); if depth == 0 { break }; continue }
                p.next()
            }
        }
        return p.parseWithTernary(&ast.IdentExpr{Pos: pos, Name: "func"}, minPrec), true
    case token.String:
        v := p.cur.Lexeme
        pos := p.cur.Pos
        // strip quotes
        if len(v) >= 2 {
            v = v[1 : len(v)-1]
        }
        p.next()
        left := ast.Expr(&ast.StringLit{Pos: pos, Value: v})
        return p.parseWithTernary(left, minPrec), true
    case token.Number:
        t := p.cur.Lexeme
        pos := p.cur.Pos
        p.next()
        left := ast.Expr(&ast.NumberLit{Pos: pos, Text: t})
        return p.parseWithTernary(left, minPrec), true
    case token.DurationLit:
        t := p.cur.Lexeme
        pos := p.cur.Pos
        p.next()
        left := ast.Expr(&ast.DurationLit{Pos: pos, Text: t})
        return p.parseWithTernary(left, minPrec), true
    default:
        return nil, false
    }
}
