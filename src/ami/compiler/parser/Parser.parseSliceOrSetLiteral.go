package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseSliceOrSetLiteral parses either a slice or set literal after seeing the name and a '<'.
func (p *Parser) parseSliceOrSetLiteral(isSlice bool, namePos source.Position) (ast.Expr, bool) {
    // consume '<'
    p.next()
    if !p.isTypeName(p.cur.Kind) {
        p.errf("expected type name after '<', got %q", p.cur.Lexeme)
        return nil, false
    }
    tname := p.cur.Lexeme
    p.next()
    // Support qualified element types like pkg.Type
    tname = p.captureQualifiedType(tname)
    if p.cur.Kind != token.Gt {
        p.errf("expected '>' after type name, got %q", p.cur.Lexeme)
        return nil, false
    }
    p.next()
    if p.cur.Kind != token.LBraceSym {
        p.errf("expected '{' to start literal, got %q", p.cur.Lexeme)
        return nil, false
    }
    lb := p.cur.Pos
    p.next()
    var elems []ast.Expr
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        e, ok := p.parseExprPrec(1)
        if ok {
            elems = append(elems, e)
        } else {
            p.errf("unexpected token in literal: %q", p.cur.Lexeme)
            p.syncUntil(token.CommaSym, token.RBraceSym)
        }
        if p.cur.Kind == token.CommaSym {
            p.next()
            continue
        }
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym {
        p.next()
    } else {
        p.errf("missing '}' in literal")
    }
    if isSlice {
        return &ast.SliceLit{Pos: namePos, TypeName: tname, LBrace: lb, Elems: elems, RBrace: rb}, true
    }
    return &ast.SetLit{Pos: namePos, TypeName: tname, LBrace: lb, Elems: elems, RBrace: rb}, true
}
