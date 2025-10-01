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

