package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseWithTernary continues parsing binary RHS with precedence and then
// recognizes the conditional operator: cond ? then : else. The ternary has
// the lowest precedence and is right-associative.
func (p *Parser) parseWithTernary(left ast.Expr, minPrec int) ast.Expr {
    // First, consume any binary operators per precedence table.
    left = p.parseBinaryRHS(left, minPrec)
    // Then, check for conditional operator.
    if p.cur.Kind == token.QuestionSym {
        // consume '?'
        p.next()
        // parse 'then' expression
        thenExpr, ok := p.parseExprPrec(1)
        if !ok {
            return left
        }
        // expect ':'
        if p.cur.Kind != token.ColonSym {
            return left
        }
        p.next()
        elseExpr, ok := p.parseExprPrec(1)
        if !ok {
            return left
        }
        return &ast.ConditionalExpr{Pos: ePos(left), Cond: left, Then: thenExpr, Else: elseExpr}
    }
    return left
}

