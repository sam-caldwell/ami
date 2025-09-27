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
    // collected leading comments to attach to the next node
    pending []ast.Comment
}

// New creates a new Parser bound to the provided file.
func New(f *source.File) *Parser {
    p := &Parser{s: scanner.New(f)}
    p.advance()
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
        im := &ast.ImportDecl{Pos: pos, Path: path, Leading: p.pending}
        p.pending = nil
        f.Decls = append(f.Decls, im)
        p.next()
    }

    // zero or more function declarations: `func Name(params) [results] {}` scaffold
    for p.cur.Kind == token.KwFunc {
        fn, err := p.parseFuncDecl()
        if err != nil { return nil, err }
        f.Decls = append(f.Decls, fn)
    }

    // zero or more pipelines: `pipeline Name() {}` scaffold
    for p.cur.Kind == token.KwPipeline {
        pd, err := p.parsePipelineDecl()
        if err != nil { return nil, err }
        f.Decls = append(f.Decls, pd)
    }

    // optional top-level error block: `error {}` scaffold
    for p.cur.Kind == token.KwError {
        eb, err := p.parseErrorBlock()
        if err != nil { return nil, err }
        f.Decls = append(f.Decls, eb)
    }
    return f, nil
}

func (p *Parser) next() { p.advance() }

// advance reads next token, collecting comments into pending and stopping at the
// first non-comment token (or EOF).
func (p *Parser) advance() {
    for {
        p.cur = p.s.Next()
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            p.pending = append(p.pending, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            continue
        }
        return
    }
}

// parseFuncDecl parses a function declaration with optional params and result tuple.
func (p *Parser) parseFuncDecl() (*ast.FuncDecl, error) {
    pos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident {
        return nil, fmt.Errorf("expected function name, got %q", p.cur.Lexeme)
    }
    name := p.cur.Lexeme
    p.next()
    params, err := p.parseParamList()
    if err != nil { return nil, err }
    results, err := p.parseResultList()
    if err != nil { return nil, err }
    body, err := p.parseBlock()
    if err != nil { return nil, err }
    fn := &ast.FuncDecl{Pos: pos, Name: name, Params: params, Results: results, Body: body, Leading: p.pending}
    p.pending = nil
    return fn, nil
}

func (p *Parser) parseParamList() ([]ast.Param, error) {
    if p.cur.Kind != token.LParenSym { return nil, fmt.Errorf("expected '(', got %q", p.cur.Lexeme) }
    p.next()
    var params []ast.Param
    for p.cur.Kind != token.RParenSym {
        if p.cur.Kind != token.Ident { return nil, fmt.Errorf("expected param name, got %q", p.cur.Lexeme) }
        nameTok := p.cur
        p.next()
        var typ string
        if p.cur.Kind == token.Ident { // treat following ident as type scaffold
            typ = p.cur.Lexeme
            p.next()
        }
        params = append(params, ast.Param{Name: nameTok.Lexeme, Pos: nameTok.Pos, Type: typ})
        if p.cur.Kind == token.CommaSym { p.next(); continue }
        if p.cur.Kind != token.RParenSym { return nil, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lexeme) }
    }
    // consume ')'
    p.next()
    return params, nil
}

func (p *Parser) parseResultList() ([]ast.Result, error) {
    // Optional tuple of results in parentheses
    if p.cur.Kind != token.LParenSym { return nil, nil }
    p.next()
    var results []ast.Result
    for p.cur.Kind != token.RParenSym {
        if p.cur.Kind != token.Ident { return nil, fmt.Errorf("expected result ident, got %q", p.cur.Lexeme) }
        results = append(results, ast.Result{Pos: p.cur.Pos, Type: p.cur.Lexeme})
        p.next()
        if p.cur.Kind == token.CommaSym { p.next(); continue }
        if p.cur.Kind != token.RParenSym { return nil, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lexeme) }
    }
    p.next()
    return results, nil
}

func (p *Parser) parseBlock() (*ast.BlockStmt, error) {
    if p.cur.Kind != token.LBraceSym { return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme) }
    lb := p.cur.Pos
    // Tolerant block: consume until matching '}' with simple depth counter.
    depth := 0
    for {
        if p.cur.Kind == token.LBraceSym { depth++ }
        p.next()
        if p.cur.Kind == token.RBraceSym {
            depth--
            if depth == 0 {
                rb := p.cur.Pos
                p.next()
                return &ast.BlockStmt{LBrace: lb, RBrace: rb}, nil
            }
        }
        // stop at EOF
        if p.cur.Kind == token.EOF { return &ast.BlockStmt{LBrace: lb}, nil }
    }
}

func (p *Parser) parsePipelineDecl() (*ast.PipelineDecl, error) {
    pos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident { return nil, fmt.Errorf("expected pipeline name, got %q", p.cur.Lexeme) }
    name := p.cur.Lexeme
    p.next()
    // consume empty param list for now
    if p.cur.Kind != token.LParenSym { return nil, fmt.Errorf("expected '(', got %q", p.cur.Lexeme) }
    p.next()
    if p.cur.Kind != token.RParenSym { return nil, fmt.Errorf("expected ')', got %q", p.cur.Lexeme) }
    p.next()
    // parse body with optional leading error block association
    if p.cur.Kind != token.LBraceSym { return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme) }
    // consume '{'
    lb := p.cur.Pos
    p.next()
    var errblk *ast.ErrorBlock
    if p.cur.Kind == token.KwError {
        eb, err := p.parseErrorBlock()
        if err != nil { return nil, err }
        errblk = eb
    }
    // expect closing '}'
    if p.cur.Kind != token.RBraceSym { return nil, fmt.Errorf("expected '}', got %q", p.cur.Lexeme) }
    rb := p.cur.Pos
    p.next()
    pd := &ast.PipelineDecl{Pos: pos, Name: name, Body: &ast.BlockStmt{LBrace: lb, RBrace: rb}, Error: errblk, Leading: p.pending}
    p.pending = nil
    return pd, nil
}

func (p *Parser) parseErrorBlock() (*ast.ErrorBlock, error) {
    pos := p.cur.Pos
    p.next()
    body, err := p.parseBlock()
    if err != nil { return nil, err }
    eb := &ast.ErrorBlock{Pos: pos, Body: body, Leading: p.pending}
    p.pending = nil
    return eb, nil
}
