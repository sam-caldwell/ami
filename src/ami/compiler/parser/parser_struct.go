package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    scan "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Parser turns tokens into a lightweight AST.
type Parser struct {
    s      *scan.Scanner
    cur    tok.Token
    file   string
    errors []diag.Diagnostic
}

// New returns a new Parser over the given source string.
func New(src string) *Parser {
    p := &Parser{s: scan.New(src), file: "input.ami"}
    p.next()
    return p
}

// Errors returns collected diagnostics from tolerant parsing.
func (p *Parser) Errors() []diag.Diagnostic { return append([]diag.Diagnostic(nil), p.errors...) }

func (p *Parser) next() { p.cur = p.s.Next() }

func (p *Parser) posFrom(t tok.Token) astpkg.Position {
    return astpkg.Position{Line: t.Line, Column: t.Column, Offset: t.Offset}
}

func (p *Parser) consumeComments() []astpkg.Comment {
    cs := p.s.ConsumeComments()
    if len(cs) == 0 {
        return nil
    }
    out := make([]astpkg.Comment, 0, len(cs))
    for _, c := range cs {
        out = append(out, astpkg.Comment{Text: c.Text, Pos: astpkg.Position{Line: c.Line, Column: c.Column, Offset: c.Offset}})
    }
    return out
}

func (p *Parser) errorf(msg string) {
    p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PARSE", Message: msg, File: p.file})
}

// synchronize advances the parser until a likely statement boundary is found.
func (p *Parser) synchronize() {
    for p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.SEMI {
            p.next()
            return
        }
        switch p.cur.Kind {
        case tok.KW_FUNC, tok.KW_IMPORT, tok.KW_PIPELINE, tok.KW_PACKAGE, tok.RBRACE:
            return
        }
        p.next()
    }
}
