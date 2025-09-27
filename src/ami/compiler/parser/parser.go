package parser

import (
    "fmt"
    "strings"

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
    f.PackagePos = p.cur.Pos
    p.next()

    // zero or more imports: `import ident`
    for p.cur.Kind == token.KwImport {
        startPos := p.cur.Pos
        p.next()
        if p.cur.Kind == token.LParenSym {
            // block form: import ( line... )
            p.next()
            for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                if p.cur.Kind == token.SemiSym || p.cur.Kind == token.CommaSym {
                    p.next(); continue
                }
                if p.cur.Kind != token.Ident && p.cur.Kind != token.String {
                    p.errf("expected import path in block, got %q", p.cur.Lexeme)
                    p.next()
                    continue
                }
                alias := ""
                aliasPos := p.cur.Pos
                var path string
                var ppos source.Position
                if p.cur.Kind == token.String {
                    path = p.cur.Lexeme
                    ppos = p.cur.Pos
                    if len(path) >= 2 { path = path[1:len(path)-1] }
                    p.next()
                } else if p.cur.Kind == token.Ident {
                    // Possible alias form: alias "path"
                    alias = p.cur.Lexeme
                    aliasPos = p.cur.Pos
                    p.next()
                    if p.cur.Kind == token.String {
                        path = p.cur.Lexeme
                        ppos = p.cur.Pos
                        if len(path) >= 2 { path = path[1:len(path)-1] }
                        p.next()
                    } else {
                        // Treat previous ident as path (no alias)
                        path = alias
                        ppos = aliasPos
                        alias = ""
                    }
                }
                constraint := p.parseImportConstraint()
                im := &ast.ImportDecl{Pos: startPos, Path: path, Leading: p.pending, PathPos: ppos, Alias: alias, AliasPos: aliasPos, Constraint: constraint}
                p.pending = nil
                f.Decls = append(f.Decls, im)
            }
            if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' to close import block") }
            continue
        }
        // single-line form
        if p.cur.Kind != token.Ident && p.cur.Kind != token.String {
            p.errf("expected import path, got %q", p.cur.Lexeme)
            p.syncTop()
            continue
        }
        alias := ""
        aliasPos := p.cur.Pos
        var path string
        var ppos source.Position
        if p.cur.Kind == token.String {
            path = p.cur.Lexeme
            ppos = p.cur.Pos
            if len(path) >= 2 { path = path[1:len(path)-1] }
            p.next()
        } else if p.cur.Kind == token.Ident {
            // Possible alias form
            alias = p.cur.Lexeme
            aliasPos = p.cur.Pos
            p.next()
            if p.cur.Kind == token.String {
                path = p.cur.Lexeme
                ppos = p.cur.Pos
                if len(path) >= 2 { path = path[1:len(path)-1] }
                p.next()
            } else {
                path = alias
                ppos = aliasPos
                alias = ""
            }
        }
        constraint := p.parseImportConstraint()
        im := &ast.ImportDecl{Pos: startPos, Path: path, Leading: p.pending, PathPos: ppos, Alias: alias, AliasPos: aliasPos, Constraint: constraint}
        p.pending = nil
        f.Decls = append(f.Decls, im)
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
    // collect pragmas from raw file content (lines starting with '#pragma ')
    f.Pragmas = p.collectPragmas()
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
    // Optional type parameters: <T[, U [constraint]]>
    var typeParams []ast.TypeParam
    if p.cur.Kind == token.Lt {
        p.next()
        for p.cur.Kind != token.Gt && p.cur.Kind != token.EOF {
            if p.cur.Kind == token.CommaSym { p.next(); continue }
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
            if p.cur.Kind == token.CommaSym { p.next(); continue }
        }
        if p.cur.Kind == token.Gt { p.next() } else { p.errf("missing '>' to close type parameter list") }
    }
    params, lp, rp, err := p.parseParamList()
    if err != nil { return nil, err }
    results, rlp, rrp, err := p.parseResultList()
    if err != nil { return nil, err }
    body, err := p.parseFuncBlock()
    if err != nil { return nil, err }
    fn := &ast.FuncDecl{Pos: pos, NamePos: namePos, Name: name, TypeParams: typeParams, Params: params, Results: results, Body: body, Leading: p.pending,
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
        if p.isTypeName(p.cur.Kind) {
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
        if !p.isTypeName(p.cur.Kind) { return nil, lp, source.Position{}, fmt.Errorf("expected result ident, got %q", p.cur.Lexeme) }
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
            leading := p.pending; p.pending = nil
            pos := p.cur.Pos
            p.next()
            if p.cur.Kind != token.Ident { p.errf("expected var name, got %q", p.cur.Lexeme); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue }
            namePos := p.cur.Pos
            name := p.cur.Lexeme
            p.next()
            var tname string
            var tpos source.Position
            if p.isTypeName(p.cur.Kind) {
                tname = p.cur.Lexeme
                tpos = p.cur.Pos
                p.next()
            }
            var init ast.Expr
            if p.cur.Kind == token.Assign { p.next(); e, ok := p.parseExpr(); if ok { init = e } }
            vd := &ast.VarDecl{Pos: pos, Name: name, NamePos: namePos, Type: tname, TypePos: tpos, Init: init, Leading: leading}
            stmts = append(stmts, vd)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.KwDefer:
            leading := p.pending; p.pending = nil
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
            ds := &ast.DeferStmt{Pos: dpos, Call: ce, Leading: leading}
            stmts = append(stmts, ds)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.Ident:
            leading := p.pending; p.pending = nil
            // Possible assignment or general expression starting with an identifier.
            name := p.cur.Lexeme
            npos := p.cur.Pos
            p.next()
            if p.cur.Kind == token.Assign {
                p.next()
                rhs, ok := p.parseExprPrec(1)
                if !ok { p.errf("expected expression on right-hand side of assignment") }
                as := &ast.AssignStmt{Pos: npos, Name: name, NamePos: npos, Value: rhs, Leading: leading}
                stmts = append(stmts, as)
                if p.cur.Kind == token.SemiSym { p.next() }
                continue
            }
            left := p.parseIdentExpr(name, npos)
            expr := p.parseBinaryRHS(left, 1)
            es := &ast.ExprStmt{Pos: ePos(expr), X: expr, Leading: leading}
            stmts = append(stmts, es)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.String, token.Number:
            leading := p.pending; p.pending = nil
            e, ok := p.parseExprPrec(1)
            if !ok { p.errf("unexpected token in statement: %q", p.cur.Lexeme); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue }
            es := &ast.ExprStmt{Pos: ePos(e), X: e, Leading: leading}
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
    case token.Ident, token.KwSlice, token.KwSet, token.KwMap:
        name := p.cur.Lexeme
        npos := p.cur.Pos
        p.next()
        left := p.parseIdentExpr(name, npos)
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
    case *ast.BinaryExpr:
        return v.Pos
    case *ast.SliceLit:
        return v.Pos
    case *ast.SetLit:
        return v.Pos
    case *ast.MapLit:
        return v.Pos
    case *ast.SelectorExpr:
        return v.Pos
    default:
        return source.Position{}
    }
}

// parseIdentExpr parses an identifier-led expression where the initial ident
// has already been consumed. It supports dotted selector chains, container
// literals, and calls.
func (p *Parser) parseIdentExpr(first string, firstPos source.Position) ast.Expr {
    // gather dotted name parts
    parts := []string{first}
    poses := []source.Position{firstPos}
    for p.cur.Kind == token.DotSym {
        p.next()
        if p.cur.Kind != token.Ident { p.errf("expected ident after '.', got %q", p.cur.Lexeme); break }
        parts = append(parts, p.cur.Lexeme)
        poses = append(poses, p.cur.Pos)
        p.next()
    }
    base := parts[0]
    // container literals after a bare keyword-like name
    if p.cur.Kind == token.Lt {
        switch base {
        case "slice":
            if lit, ok := p.parseSliceOrSetLiteral(true, firstPos); ok { return lit }
        case "set":
            if lit, ok := p.parseSliceOrSetLiteral(false, firstPos); ok { return lit }
        case "map":
            if lit, ok := p.parseMapLiteral(firstPos); ok { return lit }
        }
    }
    if p.cur.Kind == token.LParenSym {
        // join dotted parts for call names
        full := parts[0]
        for i := 1; i < len(parts); i++ { full += "." + parts[i] }
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
        return &ast.CallExpr{Pos: firstPos, Name: full, NamePos: firstPos, LParen: lp, Args: args, RParen: rp}
    }
    if len(parts) == 1 {
        return &ast.IdentExpr{Pos: firstPos, Name: first}
    }
    // selector chain
    x := ast.Expr(&ast.IdentExpr{Pos: poses[0], Name: parts[0]})
    for i := 1; i < len(parts); i++ {
        x = &ast.SelectorExpr{Pos: poses[0], X: x, Sel: parts[i], SelPos: poses[i]}
    }
    return x
}

// exprText produces a simple string representation of an expression suitable
// for debug displays (attribute args, etc.). It is not a full pretty-printer.
func exprText(e ast.Expr) string {
    switch v := e.(type) {
    case *ast.IdentExpr:
        return v.Name
    case *ast.StringLit:
        return v.Value
    case *ast.NumberLit:
        return v.Text
    case *ast.SelectorExpr:
        left := exprText(v.X)
        if left == "" { left = "?" }
        return left + "." + v.Sel
    case *ast.CallExpr:
        // return callee name with parentheses; include ellipsis when args present
        if len(v.Args) > 0 {
            return v.Name + "(…)"
        }
        return v.Name + "()"
    case *ast.SliceLit:
        return "slice"
    case *ast.SetLit:
        return "set"
    case *ast.MapLit:
        return "map"
    default:
        return ""
    }
}

// parseSliceOrSetLiteral parses either a slice or set literal after seeing the name and a '<'.
func (p *Parser) parseSliceOrSetLiteral(isSlice bool, namePos source.Position) (ast.Expr, bool) {
    // consume '<'
    p.next()
    if !p.isTypeName(p.cur.Kind) { p.errf("expected type name after '<', got %q", p.cur.Lexeme); return nil, false }
    tname := p.cur.Lexeme
    p.next()
    if p.cur.Kind != token.Gt { p.errf("expected '>' after type name, got %q", p.cur.Lexeme); return nil, false }
    p.next()
    if p.cur.Kind != token.LBraceSym { p.errf("expected '{' to start literal, got %q", p.cur.Lexeme); return nil, false }
    lb := p.cur.Pos
    p.next()
    var elems []ast.Expr
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        e, ok := p.parseExprPrec(1)
        if ok { elems = append(elems, e) } else { p.errf("unexpected token in literal: %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RBraceSym) }
        if p.cur.Kind == token.CommaSym { p.next(); continue }
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym { p.next() } else { p.errf("missing '}' in literal") }
    if isSlice {
        return &ast.SliceLit{Pos: namePos, TypeName: tname, LBrace: lb, Elems: elems, RBrace: rb}, true
    }
    return &ast.SetLit{Pos: namePos, TypeName: tname, LBrace: lb, Elems: elems, RBrace: rb}, true
}

func (p *Parser) parseMapLiteral(namePos source.Position) (ast.Expr, bool) {
    // consume '<'
    p.next()
    if !p.isTypeName(p.cur.Kind) { p.errf("expected key type after '<', got %q", p.cur.Lexeme); return nil, false }
    k := p.cur.Lexeme
    p.next()
    if p.cur.Kind != token.CommaSym { p.errf("expected ',' between key and value type, got %q", p.cur.Lexeme); return nil, false }
    p.next()
    if !p.isTypeName(p.cur.Kind) { p.errf("expected value type name, got %q", p.cur.Lexeme); return nil, false }
    v := p.cur.Lexeme
    p.next()
    if p.cur.Kind != token.Gt { p.errf("expected '>' after map type params, got %q", p.cur.Lexeme); return nil, false }
    p.next()
    if p.cur.Kind != token.LBraceSym { p.errf("expected '{' to start map literal, got %q", p.cur.Lexeme); return nil, false }
    lb := p.cur.Pos
    p.next()
    var elems []ast.MapElem
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        key, ok := p.parseExprPrec(1)
        if !ok { p.errf("expected key expression, got %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RBraceSym); if p.cur.Kind == token.CommaSym { p.next() }; continue }
        if p.cur.Kind != token.ColonSym { p.errf("expected ':', got %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RBraceSym); if p.cur.Kind == token.CommaSym { p.next() }; continue }
        p.next()
        val, ok := p.parseExprPrec(1)
        if !ok { p.errf("expected value expression, got %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RBraceSym); if p.cur.Kind == token.CommaSym { p.next() }; continue }
        elems = append(elems, ast.MapElem{Key: key, Val: val})
        if p.cur.Kind == token.CommaSym { p.next(); continue }
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym { p.next() } else { p.errf("missing '}' in map literal") }
    return &ast.MapLit{Pos: namePos, KeyType: k, ValType: v, LBrace: lb, Elems: elems, RBrace: rb}, true
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
        if p.cur.Kind == token.Ident || p.cur.Kind == token.KwIngress || p.cur.Kind == token.KwEgress ||
            p.cur.Kind == token.KwNodeTransform || p.cur.Kind == token.KwNodeFanout || p.cur.Kind == token.KwNodeCollect || p.cur.Kind == token.KwNodeMutable {
            // Determine if this is an edge or a step call
            name := p.cur.Lexeme
            namePos := p.cur.Pos
            p.next()
            // support dotted step names like io.Read
            for p.cur.Kind == token.DotSym {
                p.next()
                if p.cur.Kind != token.Ident { p.errf("expected identifier after '.', got %q", p.cur.Lexeme); break }
                name = name + "." + p.cur.Lexeme
                p.next()
            }
            if p.cur.Kind == token.Arrow {
                // Edge: name -> ident
                p.next()
                if !(p.cur.Kind == token.Ident || p.cur.Kind == token.KwNodeTransform || p.cur.Kind == token.KwNodeFanout || p.cur.Kind == token.KwNodeCollect || p.cur.Kind == token.KwNodeMutable || p.cur.Kind == token.KwIngress || p.cur.Kind == token.KwEgress) {
                    p.errf("expected identifier after '->', got %q", p.cur.Lexeme); p.syncUntil(token.SemiSym, token.RBraceSym); if p.cur.Kind == token.SemiSym { p.next() }; continue
                }
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
            // optional attributes list: Attr or dotted Attr.name(args), separated by commas
            var attrs []ast.Attr
            for p.cur.Kind == token.Ident || p.cur.Kind == token.KwType {
                aname := p.cur.Lexeme
                apos := p.cur.Pos
                p.next()
                for p.cur.Kind == token.DotSym {
                    p.next()
                    if p.cur.Kind != token.Ident { p.errf("expected ident after '.' in attribute name, got %q", p.cur.Lexeme); break }
                    aname = aname + "." + p.cur.Lexeme
                    p.next()
                }
                var aargs []ast.Arg
                if p.cur.Kind == token.LParenSym {
                    p.next()
                    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                        e, ok := p.parseExprPrec(1)
                        if ok {
                            aargs = append(aargs, ast.Arg{Pos: ePos(e), Text: exprText(e)})
                        } else {
                            p.errf("unexpected token in attr args: %q", p.cur.Lexeme)
                            p.syncUntil(token.CommaSym, token.RParenSym)
                        }
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
    body, err := p.parseStepBlock()
    if err != nil { return nil, err }
    eb := &ast.ErrorBlock{Pos: pos, Body: body, Leading: p.pending}
    p.pending = nil
    return eb, nil
}

// parseStepBlock parses a block of pipeline-like step statements into a BlockStmt
// with its Stmts populated by StepStmt nodes, tolerantly until the matching '}'.
func (p *Parser) parseStepBlock() (*ast.BlockStmt, error) {
    if p.cur.Kind != token.LBraceSym { return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme) }
    lb := p.cur.Pos
    p.next()
    var stmts []ast.Stmt
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            p.pending = append(p.pending, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            p.next(); continue
        }
        if p.cur.Kind == token.Ident || p.cur.Kind == token.KwIngress || p.cur.Kind == token.KwEgress ||
            p.cur.Kind == token.KwNodeTransform || p.cur.Kind == token.KwNodeFanout || p.cur.Kind == token.KwNodeCollect || p.cur.Kind == token.KwNodeMutable {
            // step name
            st := &ast.StepStmt{Pos: p.cur.Pos, Name: p.cur.Lexeme, Leading: p.pending}
            p.pending = nil
            p.next()
            // optional type(attr) or call args
            if p.cur.Kind == token.LParenSym {
                // Parse arguments similar to call syntax and store as ast.Arg
                p.next()
                var args []ast.Arg
                for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                    e, ok := p.parseExprPrec(1)
                    if ok {
                        args = append(args, ast.Arg{Pos: ePos(e), Text: exprText(e)})
                    } else {
                        p.errf("unexpected token in args: %q", p.cur.Lexeme)
                        p.syncUntil(token.CommaSym, token.RParenSym)
                    }
                    if p.cur.Kind == token.CommaSym { p.next(); continue }
                }
                if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' in step args") }
                st.Args = args
            }
            // attributes list (dotted names with optional parens)
            var attrs []ast.Attr
            for p.cur.Kind == token.Ident || p.cur.Kind == token.KwType {
                aname := p.cur.Lexeme
                apos := p.cur.Pos
                p.next()
                for p.cur.Kind == token.DotSym { p.next(); if p.cur.Kind != token.Ident { p.errf("expected ident after '.' in attribute name, got %q", p.cur.Lexeme); break }; aname += "." + p.cur.Lexeme; p.next() }
                var aargs []ast.Arg
                if p.cur.Kind == token.LParenSym {
                    p.next()
                    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                        e, ok := p.parseExprPrec(1)
                        if ok { aargs = append(aargs, ast.Arg{Pos: ePos(e), Text: exprText(e)}) } else { p.errf("unexpected token in attr args: %q", p.cur.Lexeme); p.syncUntil(token.CommaSym, token.RParenSym) }
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
        // unknown tokens: recover to semicolon or '}'
        p.errf("unexpected token in error block: %q", p.cur.Lexeme)
        p.syncUntil(token.SemiSym, token.RBraceSym)
        if p.cur.Kind == token.SemiSym { p.next() }
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym { p.next() }
    return &ast.BlockStmt{LBrace: lb, RBrace: rb, Stmts: stmts}, nil
}

// pendingOrNew is a small helper to use or initialize p.pending safely.
func pendingOrNew(_ []ast.Comment, p *Parser) []ast.Comment { return p.pending }

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

// collectPragmas scans the raw file content for lines beginning with '#pragma '
// and returns them as AST pragmas with 1-based line positions.
func (p *Parser) collectPragmas() []ast.Pragma {
    if p == nil || p.s == nil { return nil }
    // reach into scanner's file via helper below
    // Fallback: use token stream won’t contain raw lines; so reparse via positions is complex.
    // We can access the file pointer through the Scanner struct directly since we’re in the same module.
    // Define a local type assertion to access p.s.file.
    var content string
    // unsafe but in-package: reference the unexported field through a helper
    content = p.sFileContent()
    if content == "" { return nil }
    var out []ast.Pragma
    line := 1
    start := 0
    for i := 0; i <= len(content); i++ {
        if i == len(content) || content[i] == '\n' {
            ln := content[start:i]
            if len(ln) >= 8 && ln[:8] == "#pragma " {
                text := ln[8:]
                pr := ast.Pragma{Pos: source.Position{Line: line, Column: 1, Offset: start}, Text: text}
                // parse schema: domain:key [args...]
                // split first space to get head and rest
                head := text
                rest := ""
                if sp := indexSpace(text); sp >= 0 {
                    head = text[:sp]
                    rest = strings.TrimSpace(text[sp+1:])
                }
                if c := strings.Index(head, ":"); c >= 0 {
                    pr.Domain = head[:c]
                    pr.Key = head[c+1:]
                } else {
                    pr.Domain = head
                }
                pr.Value = rest
                if rest != "" {
                    // tokenize by spaces (no quoted parsing for now)
                    fields := strings.Fields(rest)
                    pr.Args = append(pr.Args, fields...)
                    pr.Params = map[string]string{}
                    for _, tok := range fields {
                        if eq := strings.Index(tok, "="); eq > 0 {
                            k := tok[:eq]
                            v := tok[eq+1:]
                            // strip surrounding quotes on value when present
                            if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
                                v = v[1:len(v)-1]
                            }
                            pr.Params[k] = v
                        }
                    }
                    if len(pr.Params) == 0 { pr.Params = nil }
                }
                out = append(out, pr)
            }
            line++
            start = i + 1
        }
    }
    return out
}

// sFileContent returns the underlying source file content for pragma scanning.
func (p *Parser) sFileContent() string {
    // Reach into scanner since we are within the same module.
    type hasFile interface{ FileContent() string }
    if p == nil || p.s == nil { return "" }
    // Provide a tiny adapter via a method on scanner.Scanner.
    return p.s.FileContent()
}

// indexSpace returns the index of the first ASCII space or tab, or -1.
func indexSpace(s string) int {
    for i := 0; i < len(s); i++ {
        if s[i] == ' ' || s[i] == '\t' { return i }
    }
    return -1
}

// (removed duplicate isTypeName; see richer implementation below)

// parseImportConstraint parses an optional version constraint that follows an import path.
// Supported form (scaffold): ">= vMAJOR.MINOR.PATCH[-PRERELEASE[.N]]" with or without spaces.
// Returns the canonical string (e.g., ">= v1.2.3-rc.1") or empty string when none.
func (p *Parser) parseImportConstraint() string {
    if p.cur.Kind != token.Ge {
        return ""
    }
    // capture operator
    op := p.cur.Lexeme
    if op == "" { op = ">=" }
    p.next()
    // allow quoted version in a single string token
    if p.cur.Kind == token.String {
        lex := p.cur.Lexeme
        if len(lex) >= 2 { lex = lex[1:len(lex)-1] }
        // basic validation: expect leading 'v'
        if len(lex) == 0 || lex[0] != 'v' {
            p.errf("version constraint must start with 'v', got %q", lex)
        }
        p.next()
        return op + " " + lex
    }
    // Otherwise, accumulate a tolerant SemVer string out of tokens: [Ident|Number|'.'|'-']+
    var out string
    for {
        switch p.cur.Kind {
        case token.Ident, token.Number:
            out += p.cur.Lexeme
            p.next()
        case token.DotSym:
            out += "."
            p.next()
        case token.Minus:
            out += "-"
            p.next()
        default:
            if out == "" { p.errf("expected version after operator, got %q", p.cur.Lexeme) }
            return op + " " + out
        }
    }
}

// isTypeName returns true if the current token kind is a valid type name token
// (an identifier or a recognized primitive type keyword).
func (p *Parser) isTypeName(k token.Kind) bool {
    switch k {
    case token.Ident,
        token.KwBool, token.KwByte, token.KwInt, token.KwInt8, token.KwInt16, token.KwInt32, token.KwInt64, token.KwInt128,
        token.KwUint, token.KwUint8, token.KwUint16, token.KwUint32, token.KwUint64, token.KwUint128,
        token.KwFloat32, token.KwFloat64, token.KwStringTy, token.KwRune:
        return true
    default:
        return false
    }
}
