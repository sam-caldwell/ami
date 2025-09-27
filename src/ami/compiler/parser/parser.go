package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// Parser implements a minimal recursive-descent parser.
type Parser struct {
    s   *scanner.Scanner
    cur token.Token
}

// New creates a new Parser bound to the provided file.
func New(f *source.File) *Parser {
    p := &Parser{s: scanner.New(f)}
    p.cur = p.s.Next()
    return p
}

// ParseFile parses a single file, recognizing only a package declaration.
func (p *Parser) ParseFile() (*ast.File, error) {
    if p == nil { return nil, fmt.Errorf("nil parser") }
    f := &ast.File{}
    // Expect 'package' ident
    if p.cur.Kind != token.Ident || p.cur.Lexeme != "package" {
        return nil, fmt.Errorf("expected 'package', got %q", p.cur.Lexeme)
    }
    p.next()
    if p.cur.Kind != token.Ident {
        return nil, fmt.Errorf("expected package name, got %q", p.cur.Lexeme)
    }
    f.PackageName = p.cur.Lexeme
    p.next()
    return f, nil
}

func (p *Parser) next() { p.cur = p.s.Next() }

