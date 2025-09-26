package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseParamList parses zero or more parameters until ')'
func (p *Parser) parseParamList() []astpkg.Param {
    var params []astpkg.Param
    for p.cur.Kind != tok.RPAREN && p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.COMMA { p.next(); continue }
        var name string
        if p.cur.Kind == tok.IDENT {
            ident := p.cur.Lexeme
            p.next()
            if p.cur.Kind == tok.STAR {
                p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'*' pointer type/dereference is not allowed; AMI does not expose raw pointers (see 2.3.2)", File: p.file})
                p.next()
            }
            if tr, ok := p.parseType(); ok {
                name = ident
                params = append(params, astpkg.Param{Name: name, Type: tr})
            } else {
                params = append(params, astpkg.Param{Name: "", Type: astpkg.TypeRef{Name: ident}})
            }
        } else {
            if tr, ok := p.parseType(); ok { params = append(params, astpkg.Param{Name: "", Type: tr}) } else { p.next() }
        }
        if p.cur.Kind == tok.COMMA { p.next() }
    }
    return params
}

func (p *Parser) parseResultList() []astpkg.TypeRef {
    var results []astpkg.TypeRef
    for p.cur.Kind != tok.RPAREN && p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.COMMA { p.next(); continue }
        if p.cur.Kind == tok.STAR {
            p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'*' pointer type/dereference is not allowed; AMI does not expose raw pointers (see 2.3.2)", File: p.file})
            p.next()
        }
        if tr, ok := p.parseType(); ok { results = append(results, tr) } else { p.next() }
        if p.cur.Kind == tok.COMMA { p.next() }
    }
    return results
}

// parseType parses '*'? '[]'? IDENT|KW_MAP|KW_SET|KW_SLICE ('<' Type {',' Type } '>')?
func (p *Parser) parseType() (astpkg.TypeRef, bool) {
    var tr astpkg.TypeRef
    start := p.cur.Offset
    if p.cur.Kind == tok.STAR { tr.Ptr = true; p.next(); start = p.cur.Offset }
    if p.cur.Kind == tok.LBRACK { start = p.cur.Offset; p.next(); if p.cur.Kind == tok.RBRACK { tr.Slice = true; p.next() } }
    switch p.cur.Kind {
    case tok.IDENT, tok.KW_MAP, tok.KW_SET, tok.KW_SLICE:
    default:
        return tr, false
    }
    tr.Name = p.cur.Lexeme
    tr.Offset = start
    p.next()
    if p.cur.Kind == tok.LT {
        p.next()
        for p.cur.Kind != tok.EOF {
            if p.cur.Kind == tok.GT { p.next(); break }
            if p.cur.Kind == tok.COMMA { p.next(); continue }
            if arg, ok := p.parseType(); ok { tr.Args = append(tr.Args, arg); continue }
            p.next()
        }
    }
    return tr, true
}

