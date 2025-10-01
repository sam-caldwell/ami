package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

func (p *Parser) parseExpr() (ast.Expr, bool) {
    return p.parseExprPrec(1)
}

