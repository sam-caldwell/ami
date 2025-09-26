package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseTypeParams parses an optional generic parameter list after a function name:
//   '<' IDENT [IDENT]? { ',' IDENT [IDENT]? } '>'
// The second IDENT, when present, is captured as a tolerant constraint (e.g., 'any').
func (p *Parser) parseTypeParams() []astpkg.TypeParam {
    if p.cur.Kind != tok.LT { return nil }
    var list []astpkg.TypeParam
    // consume '<'
    p.next()
    for p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.GT { p.next(); break }
        if p.cur.Kind == tok.COMMA { p.next(); continue }
        if p.cur.Kind != tok.IDENT {
            // tolerate and skip until comma or '>'
            p.next()
            continue
        }
        name := p.cur.Lexeme
        p.next()
        constraint := ""
        // optional single IDENT constraint (e.g., any)
        if p.cur.Kind == tok.IDENT {
            constraint = p.cur.Lexeme
            p.next()
        }
        list = append(list, astpkg.TypeParam{Name: name, Constraint: constraint})
        if p.cur.Kind == tok.COMMA { p.next() }
    }
    return list
}

