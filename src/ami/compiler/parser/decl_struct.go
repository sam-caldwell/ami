package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseStructDecl parses: struct IDENT '{' fields '}'
// field := IDENT TypeRef
// separators: optional comma or semicolon between fields
func (p *Parser) parseStructDecl() astpkg.StructDecl {
    p.next() // consume 'struct'
    name := ""
    if p.cur.Kind == tok.IDENT {
        name = p.cur.Lexeme
        p.next()
    } else {
        return astpkg.StructDecl{}
    }
    if p.cur.Kind != tok.LBRACE {
        return astpkg.StructDecl{}
    }
    p.next() // consume '{'
    var fields []astpkg.Field
    for p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.RBRACE {
            p.next()
            break
        }
        if p.cur.Kind == tok.COMMA || p.cur.Kind == tok.SEMI {
            p.next()
            continue
        }
        if p.cur.Kind != tok.IDENT {
            // skip token and continue
            p.next()
            continue
        }
        fname := p.cur.Lexeme
        p.next()
        ftype, ok := p.parseType()
        if !ok {
            fields = append(fields, astpkg.Field{Name: fname, Type: astpkg.TypeRef{}})
        } else {
            fields = append(fields, astpkg.Field{Name: fname, Type: ftype})
        }
        if p.cur.Kind == tok.COMMA || p.cur.Kind == tok.SEMI {
            p.next()
        }
    }
    return astpkg.StructDecl{Name: name, Fields: fields}
}

