package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseMapLiteral(namePos source.Position) (ast.Expr, bool) {
    // consume '<'
    p.next()
    if !p.isTypeName(p.cur.Kind) {
        p.errf("expected key type after '<', got %q", p.cur.Lexeme)
        return nil, false
    }
    k := p.cur.Lexeme
    p.next()
    // Support qualified key types like pkg.Type
    k = p.captureQualifiedType(k)
    if p.cur.Kind != token.CommaSym {
        p.errf("expected ',' between key and value type, got %q", p.cur.Lexeme)
        return nil, false
    }
    p.next()
    if !p.isTypeName(p.cur.Kind) {
        p.errf("expected value type name, got %q", p.cur.Lexeme)
        return nil, false
    }
    v := p.cur.Lexeme
    p.next()
    // Support qualified value types
    v = p.captureQualifiedType(v)
    if p.cur.Kind != token.Gt {
        p.errf("expected '>' after map type params, got %q", p.cur.Lexeme)
        return nil, false
    }
    p.next()
    if p.cur.Kind != token.LBraceSym {
        p.errf("expected '{' to start map literal, got %q", p.cur.Lexeme)
        return nil, false
    }
    lb := p.cur.Pos
    p.next()
    var elems []ast.MapElem
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        key, ok := p.parseExprPrec(1)
        if !ok {
            p.errf("expected key expression, got %q", p.cur.Lexeme)
            p.syncUntil(token.CommaSym, token.RBraceSym)
            if p.cur.Kind == token.CommaSym {
                p.next()
            }
            continue
        }
        if p.cur.Kind != token.ColonSym {
            p.errf("expected ':', got %q", p.cur.Lexeme)
            p.syncUntil(token.CommaSym, token.RBraceSym)
            if p.cur.Kind == token.CommaSym {
                p.next()
            }
            continue
        }
        p.next()
        val, ok := p.parseExprPrec(1)
        if !ok {
            p.errf("expected value expression, got %q", p.cur.Lexeme)
            p.syncUntil(token.CommaSym, token.RBraceSym)
            if p.cur.Kind == token.CommaSym {
                p.next()
            }
            continue
        }
        elems = append(elems, ast.MapElem{Key: key, Val: val})
        if p.cur.Kind == token.CommaSym {
            p.next()
            continue
        }
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym {
        p.next()
    } else {
        p.errf("missing '}' in map literal")
    }
    return &ast.MapLit{Pos: namePos, KeyType: k, ValType: v, LBrace: lb, Elems: elems, RBrace: rb}, true
}
