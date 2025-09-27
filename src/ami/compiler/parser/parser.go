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
    errors  []error
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
        p.errf("expected 'package', got %q", p.cur.Lexeme)
        p.syncTop()
        if p.cur.Kind != token.KwPackage { return f, p.firstErr() }
    }
    p.next()
    if p.cur.Kind != token.Ident {
        p.errf("expected package name, got %q", p.cur.Lexeme)
        p.syncTop()
        if p.cur.Kind != token.Ident { return f, p.firstErr() }
    }
    f.PackageName = p.cur.Lexeme
    p.next()

    // zero or more imports: `import ident`
    for p.cur.Kind == token.KwImport {
        pos := p.cur.Pos
        p.next()
        if p.cur.Kind != token.Ident && p.cur.Kind != token.String {
            p.errf("expected import path, got %q", p.cur.Lexeme)
            p.syncTop()
            continue
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
        if err != nil {
            p.errf("%v", err)
            p.syncTop()
        } else {
            f.Decls = append(f.Decls, fn)
        }
    }

    // zero or more pipelines: `pipeline Name() {}` scaffold
    for p.cur.Kind == token.KwPipeline {
        pd, err := p.parsePipelineDecl()
        if err != nil {
            p.errf("%v", err)
            p.syncTop()
        } else {
            f.Decls = append(f.Decls, pd)
        }
    }

    // optional top-level error block: `error {}` scaffold
    for p.cur.Kind == token.KwError {
        eb, err := p.parseErrorBlock()
        if err != nil {
            p.errf("%v", err)
            p.syncTop()
        } else {
            f.Decls = append(f.Decls, eb)
        }
    }
    return f, p.firstErr()
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
    namePos := p.cur.Pos
    p.next()
    params, lp, rp, err := p.parseParamList()
    if err != nil { return nil, err }
    results, rlp, rrp, err := p.parseResultList()
    if err != nil { return nil, err }
    body, err := p.parseFuncBlock()
    if err != nil { return nil, err }
    fn := &ast.FuncDecl{Pos: pos, NamePos: namePos, Name: name, Params: params, Results: results, Body: body, Leading: p.pending,
        ParamsLParen: lp, ParamsRParen: rp, ResultsLParen: rlp, ResultsRParen: rrp}
    p.pending = nil
    return fn, nil
}

func (p *Parser) parseParamList() ([]ast.Param, source.Position, source.Position, error) {
    if p.cur.Kind != token.LParenSym { return nil, source.Position{}, source.Position{}, fmt.Errorf("expected '(', got %q", p.cur.Lexeme) }
    lp := p.cur.Pos
    p.next()
    var params []ast.Param
    var pend []ast.Comment
    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            pend = append(pend, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            p.next()
            continue
        }
        if p.cur.Kind != token.Ident { return nil, lp, source.Position{}, fmt.Errorf("expected param name, got %q", p.cur.Lexeme) }
        nameTok := p.cur
        p.next()
        var typ string
        if p.cur.Kind == token.Ident { // treat following ident as type scaffold
            typ = p.cur.Lexeme
            p.next()
        }
        params = append(params, ast.Param{Name: nameTok.Lexeme, Pos: nameTok.Pos, Type: typ, Leading: pend})
        pend = nil
        if p.cur.Kind == token.CommaSym { p.next(); continue }
        if p.cur.Kind != token.RParenSym { return nil, lp, source.Position{}, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lexeme) }
    }
    rp := p.cur.Pos
    // consume ')'
    if p.cur.Kind == token.RParenSym { p.next() }
    return params, lp, rp, nil
}

func (p *Parser) parseResultList() ([]ast.Result, source.Position, source.Position, error) {
    // Optional tuple of results in parentheses
    if p.cur.Kind != token.LParenSym { return nil, source.Position{}, source.Position{}, nil }
    lp := p.cur.Pos
    p.next()
    var results []ast.Result
    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
        if p.cur.Kind != token.Ident { return nil, lp, source.Position{}, fmt.Errorf("expected result ident, got %q", p.cur.Lexeme) }
        results = append(results, ast.Result{Pos: p.cur.Pos, Type: p.cur.Lexeme})
        p.next()
        if p.cur.Kind == token.CommaSym { p.next(); continue }
        if p.cur.Kind != token.RParenSym { return nil, lp, source.Position{}, fmt.Errorf("expected ',' or ')', got %q", p.cur.Lexeme) }
    }
    rp := p.cur.Pos
    if p.cur.Kind == token.RParenSym { p.next() }
    return results, lp, rp, nil
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

func (p *Parser) parseFuncBlock() (*ast.BlockStmt, error) {
    if p.cur.Kind != token.LBraceSym { return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme) }
    lb := p.cur.Pos
    p.next()
    var stmts []ast.Stmt
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            p.pending = append(p.pending, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            p.next(); continue
        }
        switch p.cur.Kind {
        case token.KwReturn:
            pos := p.cur.Pos
            p.next()
            var results []ast.Expr
            if p.isExprStart(p.cur.Kind) {
                e, ok := p.parseExpr()
                if ok { results = append(results, e) }
                for p.cur.Kind == token.CommaSym { p.next(); e, ok = p.parseExpr(); if ok { results = append(results, e) } }
            }
            rs := &ast.ReturnStmt{Pos: pos, Results: results, Leading: p.pending}
            p.pending = nil
            stmts = append(stmts, rs)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.KwVar:
            pos := p.cur.Pos
            p.next()
            if p.cur.Kind != token.Ident { p.errf("expected var name, got %q", p.cur.Lexeme); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue }
            namePos := p.cur.Pos
            name := p.cur.Lexeme
            p.next()
            var tname string
            var tpos source.Position
            if p.cur.Kind == token.Ident {
                tname = p.cur.Lexeme
                tpos = p.cur.Pos
                p.next()
            }
            var init ast.Expr
            if p.cur.Kind == token.Assign { p.next(); e, ok := p.parseExpr(); if ok { init = e } }
            vd := &ast.VarDecl{Pos: pos, Name: name, NamePos: namePos, Type: tname, TypePos: tpos, Init: init, Leading: p.pending}
            p.pending = nil
            stmts = append(stmts, vd)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.KwDefer:
            dpos := p.cur.Pos
            p.next()
            callExpr, ok := p.parseExpr()
            if !ok {
                p.errf("expected call after defer, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym { p.next() }
                continue
            }
            ce, _ := callExpr.(*ast.CallExpr)
            if ce == nil { p.errf("defer requires a call expression"); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue }
            ds := &ast.DeferStmt{Pos: dpos, Call: ce, Leading: p.pending}
            p.pending = nil
            stmts = append(stmts, ds)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.Ident:
            // assignment or expression starting with ident
            name := p.cur.Lexeme
            npos := p.cur.Pos
            p.next()
            if p.cur.Kind == token.Assign {
                p.next()
                rhs, ok := p.parseExprPrec(1)
                if !ok { p.errf("expected expression on right-hand side of assignment") }
                as := &ast.AssignStmt{Pos: npos, Name: name, NamePos: npos, Value: rhs, Leading: p.pending}
                p.pending = nil
                stmts = append(stmts, as)
                if p.cur.Kind == token.SemiSym { p.next() }
                continue
            }
            // build expression starting with prior ident
            var left ast.Expr = &ast.IdentExpr{Pos: npos, Name: name}
            // call suffix
            if p.cur.Kind == token.LParenSym {
                lp := p.cur.Pos
                p.next()
                var args []ast.Expr
                for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                    e, ok := p.parseExprPrec(1)
                    if ok { args = append(args, e) } else { p.errf("unexpected token in call args: %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RParenSym) }
                    if p.cur.Kind == token.CommaSym { p.next(); continue }
                }
                rp := p.cur.Pos
                if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' in call expr") }
                left = &ast.CallExpr{Pos: npos, Name: name, NamePos: npos, LParen: lp, Args: args, RParen: rp}
            }
            // binary tail
            expr := p.parseBinaryRHS(left, 1)
            es := &ast.ExprStmt{Pos: ePos(expr), X: expr, Leading: p.pending}
            p.pending = nil
            stmts = append(stmts, es)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.String, token.Number:
            e, ok := p.parseExprPrec(1)
            if !ok { p.errf("unexpected token in statement: %q", p.cur.Lexeme); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue }
            es := &ast.ExprStmt{Pos: ePos(e), X: e, Leading: p.pending}
            p.pending = nil
            stmts = append(stmts, es)
            if p.cur.Kind == token.SemiSym { p.next() }
        default:
            p.errf("unexpected token in function body: %q", p.cur.Lexeme)
            p.syncUntil(token.SemiSym, token.RBraceSym)
            if p.cur.Kind == token.SemiSym { p.next() }
        }
    }
    if p.cur.Kind != token.RBraceSym { return nil, fmt.Errorf("expected '}', got %q", p.cur.Lexeme) }
    rb := p.cur.Pos
    p.next()
    return &ast.BlockStmt{LBrace: lb, RBrace: rb, Stmts: stmts}, nil
}

func (p *Parser) isExprStart(k token.Kind) bool {
    switch k {
    case token.Ident, token.Number, token.String:
        return true
    default:
        return false
    }
}

func (p *Parser) parseExpr() (ast.Expr, bool) {
    return p.parseExprPrec(1)
}

func (p *Parser) parseExprPrec(minPrec int) (ast.Expr, bool) {
    switch p.cur.Kind {
    case token.Ident:
        name := p.cur.Lexeme
        npos := p.cur.Pos
        p.next()
        // call?
        if p.cur.Kind == token.LParenSym {
            lp := p.cur.Pos
            p.next()
            var args []ast.Expr
            for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                e, ok := p.parseExprPrec(1)
                if ok { args = append(args, e) } else { p.errf("unexpected token in call args: %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RParenSym) }
                if p.cur.Kind == token.CommaSym { p.next(); continue }
            }
            rp := p.cur.Pos
            if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' in call expr") }
            left := ast.Expr(&ast.CallExpr{Pos: npos, Name: name, NamePos: npos, LParen: lp, Args: args, RParen: rp})
            return p.parseBinaryRHS(left, minPrec), true
        }
        left := ast.Expr(&ast.IdentExpr{Pos: npos, Name: name})
        return p.parseBinaryRHS(left, minPrec), true
    case token.String:
        v := p.cur.Lexeme
        pos := p.cur.Pos
        // strip quotes
        if len(v) >= 2 { v = v[1:len(v)-1] }
        p.next()
        left := ast.Expr(&ast.StringLit{Pos: pos, Value: v})
        return p.parseBinaryRHS(left, minPrec), true
    case token.Number:
        t := p.cur.Lexeme
        pos := p.cur.Pos
        p.next()
        left := ast.Expr(&ast.NumberLit{Pos: pos, Text: t})
        return p.parseBinaryRHS(left, minPrec), true
    default:
        return nil, false
    }
}

func (p *Parser) parseBinaryRHS(left ast.Expr, minPrec int) ast.Expr {
    for {
        prec := token.Precedence(p.cur.Kind)
        if prec < minPrec || prec == 0 {
            return left
        }
        op := p.cur.Kind
        p.next()
        // parse right-hand side with higher precedence for right node
        right, ok := p.parseExprPrec(prec + 1)
        if !ok {
            // if we cannot parse rhs, stop chaining
            return left
        }
        left = &ast.BinaryExpr{Pos: ePos(left), Op: op, X: left, Y: right}
    }
}

func ePos(e ast.Expr) source.Position {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Pos
    case *ast.StringLit:
        return v.Pos
    case *ast.NumberLit:
        return v.Pos
    case *ast.CallExpr:
        return v.Pos
    default:
        return source.Position{}
    }
}

func (p *Parser) parsePipelineDecl() (*ast.PipelineDecl, error) {
    pos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident { return nil, fmt.Errorf("expected pipeline name, got %q", p.cur.Lexeme) }
    name := p.cur.Lexeme
    namePos := p.cur.Pos
    p.next()
    // consume empty param list for now
    if p.cur.Kind != token.LParenSym { return nil, fmt.Errorf("expected '(', got %q", p.cur.Lexeme) }
    lp := p.cur.Pos
    p.next()
    if p.cur.Kind != token.RParenSym { return nil, fmt.Errorf("expected ')', got %q", p.cur.Lexeme) }
    rp := p.cur.Pos
    p.next()
    // parse body with optional leading error block and simple steps
    if p.cur.Kind != token.LBraceSym { return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme) }
    lb := p.cur.Pos
    p.next()
    var errblk *ast.ErrorBlock
    var stmts []ast.Stmt
    if p.cur.Kind == token.KwError {
        eb, err := p.parseErrorBlock()
        if err != nil { return nil, err }
        errblk = eb
    }
    // parse step statements until '}'
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            p.pending = append(p.pending, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            p.next()
            continue
        }
        if p.cur.Kind == token.Ident {
            // Determine if this is an edge or a step call
            name := p.cur.Lexeme
            namePos := p.cur.Pos
            p.next()
            if p.cur.Kind == token.Arrow {
                // Edge: name -> ident
                p.next()
                if p.cur.Kind != token.Ident { p.errf("expected identifier after '->', got %q", p.cur.Lexeme); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue }
                to := p.cur.Lexeme
                toPos := p.cur.Pos
                edge := &ast.EdgeStmt{Pos: namePos, From: name, FromPos: namePos, To: to, ToPos: toPos, Leading: p.pending}
                p.pending = nil
                stmts = append(stmts, edge)
                p.next()
                if p.cur.Kind == token.SemiSym { p.next() }
                continue
            }
            // Step call starting with name already consumed
            st := &ast.StepStmt{Pos: namePos, Name: name, Leading: p.pending}
            p.pending = nil
            // optional args
            if p.cur.Kind == token.LParenSym {
                p.next()
                var args []ast.Arg
                for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                    switch p.cur.Kind {
                    case token.Ident:
                        args = append(args, ast.Arg{Pos: p.cur.Pos, Text: p.cur.Lexeme})
                    case token.String:
                        args = append(args, ast.Arg{Pos: p.cur.Pos, Text: p.cur.Lexeme[1:len(p.cur.Lexeme)-1], IsString: true})
                    default:
                        p.errf("unexpected token in args: %q", p.cur.Lexeme)
                        p.syncUntil(token.CommaSym, token.RParenSym)
                    }
                    p.next()
                    if p.cur.Kind == token.CommaSym { p.next(); continue }
                }
                if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' in call") }
                st.Args = args
            }
            // optional attributes list: Attr or Attr(args), separated by commas
            var attrs []ast.Attr
            for p.cur.Kind == token.Ident {
                aname := p.cur.Lexeme
                apos := p.cur.Pos
                p.next()
                var aargs []ast.Arg
                if p.cur.Kind == token.LParenSym {
                    p.next()
                    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                        switch p.cur.Kind {
                        case token.Ident:
                            aargs = append(aargs, ast.Arg{Pos: p.cur.Pos, Text: p.cur.Lexeme})
                        case token.String:
                            aargs = append(aargs, ast.Arg{Pos: p.cur.Pos, Text: p.cur.Lexeme[1:len(p.cur.Lexeme)-1], IsString: true})
                        default:
                            p.errf("unexpected token in attr args: %q", p.cur.Lexeme)
                            p.syncUntil(token.CommaSym, token.RParenSym)
                        }
                        p.next()
                        if p.cur.Kind == token.CommaSym { p.next(); continue }
                    }
                    if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' in attr call") }
                }
                attrs = append(attrs, ast.Attr{Pos: apos, Name: aname, Args: aargs})
                if p.cur.Kind == token.CommaSym { p.next(); continue }
            }
            st.Attrs = attrs
            stmts = append(stmts, st)
            if p.cur.Kind == token.SemiSym { p.next() }
            continue
        }
        // unknown: recover to semicolon or '}'
        p.errf("unexpected token in pipeline: %q", p.cur.Lexeme)
        p.syncUntil(token.SemiSym, token.RBraceSym)
        if p.cur.Kind == token.SemiSym { p.next() }
    }
    if p.cur.Kind != token.RBraceSym { return nil, fmt.Errorf("expected '}', got %q", p.cur.Lexeme) }
    rb := p.cur.Pos
    p.next()
    pd := &ast.PipelineDecl{Pos: pos, Name: name, NamePos: namePos, Body: &ast.BlockStmt{LBrace: lb, RBrace: rb}, Error: errblk, Leading: p.pending, Stmts: stmts, LParen: lp, RParen: rp}
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

func (p *Parser) errf(format string, args ...any) {
    p.errors = append(p.errors, fmt.Errorf(format, args...))
}

func (p *Parser) firstErr() error {
    if len(p.errors) == 0 { return nil }
    return p.errors[0]
}

// syncTop synchronizes the token stream to the next likely top-level boundary.
func (p *Parser) syncTop() {
    p.syncUntil(token.SemiSym, token.KwFunc, token.KwImport, token.KwPipeline, token.KwError, token.EOF)
}

// syncUntil advances until one of the specified kinds is found.
func (p *Parser) syncUntil(kinds ...token.Kind) {
    set := make(map[token.Kind]struct{}, len(kinds))
    for _, k := range kinds { set[k] = struct{}{} }
    for {
        if _, ok := set[p.cur.Kind]; ok { return }
        if p.cur.Kind == token.EOF { return }
        p.next()
    }
}

// ParseFileCollect parses a file and returns the file and any collected errors.
func (p *Parser) ParseFileCollect() (*ast.File, []error) {
    f, _ := p.ParseFile()
    return f, p.errors
}
