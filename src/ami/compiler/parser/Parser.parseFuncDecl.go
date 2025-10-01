package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseFuncDecl parses a function declaration with optional params and result tuple.
func (p *Parser) parseFuncDecl() (*ast.FuncDecl, error) {
    pos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident {
        return nil, fmt.Errorf("expected function name, got %q", p.cur.Lexeme)
    }
    name := p.cur.Lexeme
    namePos := p.cur.Pos
    p.next()
    // Optional type parameters: <T[, U [constraint]]>
    var typeParams []ast.TypeParam
    if p.cur.Kind == token.Lt {
        p.next()
        for p.cur.Kind != token.Gt && p.cur.Kind != token.EOF {
            if p.cur.Kind == token.CommaSym {
                p.next()
                continue
            }
            if p.cur.Kind != token.Ident {
                p.errf("expected type parameter name, got %q", p.cur.Lexeme)
                p.next()
                continue
            }
            tpName := p.cur.Lexeme
            tpNamePos := p.cur.Pos
            p.next()
            // optional constraint ident (e.g., any)
            var c string
            if p.cur.Kind == token.Ident {
                c = p.cur.Lexeme
                p.next()
            }
            typeParams = append(typeParams, ast.TypeParam{Pos: tpNamePos, Name: tpName, NamePos: tpNamePos, Constraint: c})
            if p.cur.Kind == token.CommaSym {
                p.next()
                continue
            }
        }
        if p.cur.Kind == token.Gt {
            p.next()
        } else {
            p.errf("missing '>' to close type parameter list")
        }
    }
    params, lp, rp, err := p.parseParamList()
    if err != nil {
        return nil, err
    }
    results, rlp, rrp, err := p.parseResultList()
    if err != nil {
        return nil, err
    }
    body, err := p.parseFuncBlock()
    if err != nil {
        return nil, err
    }
    fn := &ast.FuncDecl{Pos: pos, NamePos: namePos, Name: name, TypeParams: typeParams, Params: params, Results: results, Body: body, Leading: p.pending,
        Decorators: p.pendingDecos, ParamsLParen: lp, ParamsRParen: rp, ResultsLParen: rlp, ResultsRParen: rrp}
    p.pending = nil
    p.pendingDecos = nil
    return fn, nil
}
