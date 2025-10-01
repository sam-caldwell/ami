package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseEnumDecl parses: enum Name { Member (, Member)* }
func (p *Parser) parseEnumDecl() (*ast.EnumDecl, error) {
    pos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident {
        return nil, fmt.Errorf("expected enum name, got %q", p.cur.Lexeme)
    }
    name := p.cur.Lexeme
    namePos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.LBraceSym {
        return nil, fmt.Errorf("expected '{' to start enum, got %q", p.cur.Lexeme)
    }
    lb := p.cur.Pos
    p.next()
    var members []ast.EnumMember
    expectMember := true
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.CommaSym {
            // If we were expecting a member but found a comma, this is a blank member (",,")
            if expectMember {
                members = append(members, ast.EnumMember{Pos: p.cur.Pos, Name: ""})
            }
            p.next()
            expectMember = true
            continue
        }
        if p.cur.Kind != token.Ident {
            p.errf("expected enum member name, got %q", p.cur.Lexeme)
            p.syncUntil(token.CommaSym, token.RBraceSym)
            if p.cur.Kind == token.CommaSym {
                p.next()
            }
            expectMember = true
            continue
        }
        members = append(members, ast.EnumMember{Pos: p.cur.Pos, Name: p.cur.Lexeme})
        p.next()
        if p.cur.Kind == token.CommaSym {
            p.next()
            expectMember = true
            continue
        }
        expectMember = false
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym {
        p.next()
    } else {
        p.errf("missing '}' to close enum")
    }
    return &ast.EnumDecl{Pos: pos, NamePos: namePos, Name: name, LBrace: lb, Members: members, RBrace: rb}, nil
}

