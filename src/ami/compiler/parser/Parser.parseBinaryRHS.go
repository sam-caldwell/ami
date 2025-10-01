package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseBinaryRHS(left ast.Expr, minPrec int) ast.Expr {
    for {
        prec := token.Precedence(p.cur.Kind)
        if prec < minPrec || prec == 0 {
            return left
        }
        op := p.cur.Kind
        p.next()
        // parse right-hand side with higher precedence for right node
        right, ok := p.parseExprPrec(prec + 1)
        if !ok {
            // if we cannot parse rhs, stop chaining
            return left
        }
        left = &ast.BinaryExpr{Pos: ePos(left), Op: op, X: left, Y: right}
    }
}

