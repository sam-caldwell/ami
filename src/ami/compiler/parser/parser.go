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
    // Expect 'package' keyword
    if p.cur.Kind != token.KwPackage {
        return nil, fmt.Errorf("expected 'package', got %q", p.cur.Lexeme)
    }
    p.next()
    if p.cur.Kind != token.Ident {
        return nil, fmt.Errorf("expected package name, got %q", p.cur.Lexeme)
    }
    f.PackageName = p.cur.Lexeme
    p.next()

    // zero or more imports: `import ident`
    for p.cur.Kind == token.KwImport {
        pos := p.cur.Pos
        p.next()
        if p.cur.Kind != token.Ident && p.cur.Kind != token.String {
            return nil, fmt.Errorf("expected import path, got %q", p.cur.Lexeme)
        }
        path := p.cur.Lexeme
        // strip quotes if string literal
        if p.cur.Kind == token.String && len(path) >= 2 {
            path = path[1:len(path)-1]
        }
        f.Decls = append(f.Decls, &ast.ImportDecl{Pos: pos, Path: path})
        p.next()
    }

    // zero or more function declarations: `func Name() {}` scaffold
    for p.cur.Kind == token.KwFunc {
        pos := p.cur.Pos
        p.next()
        if p.cur.Kind != token.Ident {
            return nil, fmt.Errorf("expected function name, got %q", p.cur.Lexeme)
        }
        name := p.cur.Lexeme
        p.next()
        if p.cur.Kind != token.LParenSym { return nil, fmt.Errorf("expected '(', got %q", p.cur.Lexeme) }
        p.next()
        if p.cur.Kind != token.RParenSym { return nil, fmt.Errorf("expected ')', got %q", p.cur.Lexeme) }
        p.next()
        if p.cur.Kind != token.LBraceSym { return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme) }
        // consume until matching '}' minimally (no nested parsing in scaffold)
        // For now, we expect immediate '}'
        p.next()
        if p.cur.Kind != token.RBraceSym { return nil, fmt.Errorf("expected '}', got %q", p.cur.Lexeme) }
        p.next()
        f.Decls = append(f.Decls, &ast.FuncDecl{Pos: pos, Name: name})
    }
    return f, nil
}

func (p *Parser) next() { p.cur = p.s.Next() }
