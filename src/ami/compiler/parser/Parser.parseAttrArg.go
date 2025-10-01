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

