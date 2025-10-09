package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parseFuncBlock() (*ast.BlockStmt, error) {
    if p.cur.Kind != token.LBraceSym {
        return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme)
    }
    lb := p.cur.Pos
    p.next()
    var stmts []ast.Stmt
    for p.cur.Kind != token.RBraceSym && p.cur.Kind != token.EOF {
        if p.cur.Kind == token.LineComment || p.cur.Kind == token.BlockComment {
            p.pending = append(p.pending, ast.Comment{Pos: p.cur.Pos, Text: p.cur.Lexeme})
            p.next()
            continue
        }
        switch p.cur.Kind {
        case token.KwGpu:
            // Parse: gpu(attr=val, ...){ raw GPU source }
            gpos := p.cur.Pos
            p.next()
            // optional attrs in parens
            var gattrs []ast.Arg
            if p.cur.Kind == token.LParenSym {
                p.next()
                for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                    if arg, ok := p.parseAttrArg(); ok {
                        gattrs = append(gattrs, arg)
                    } else {
                        p.errf("unexpected token in gpu() attrs: %q", p.cur.Lexeme)
                        p.syncUntil(token.CommaSym, token.RParenSym)
                    }
                    if p.cur.Kind == token.CommaSym { p.next(); continue }
                }
                if p.cur.Kind == token.RParenSym { p.next() } else { p.errf("missing ')' after gpu(...) attrs") }
            }
            // require '{' and capture raw balanced content
            if p.cur.Kind != token.LBraceSym {
                p.errf("expected '{' to start gpu block, got %q", p.cur.Lexeme)
                p.syncUntil(token.RBraceSym, token.SemiSym)
                if p.cur.Kind == token.SemiSym { p.next() }
                continue
            }
            lb := p.cur.Pos
            // capture content between braces using depth tracking
            startOff := lb.Offset
            depth := 0
            var lastBraceOff int
            for {
                if p.cur.Kind == token.EOF { break }
                if p.cur.Kind == token.LBraceSym { depth++ }
                if p.cur.Kind == token.RBraceSym {
                    depth--
                    lastBraceOff = p.cur.Pos.Offset
                    p.next()
                    if depth == 0 { break }
                    continue
                }
                p.next()
            }
            src := p.s.FileContent()
            var payload string
            if lastBraceOff > startOff+1 && lastBraceOff <= len(src) {
                payload = src[startOff+1 : lastBraceOff]
            }
            g := &ast.GPUBlockStmt{Pos: gpos, Attrs: gattrs, LBrace: lb, RBrace: source.Position{Offset: lastBraceOff}, Source: payload}
            stmts = append(stmts, g)
            if p.cur.Kind == token.SemiSym { p.next() }
        case token.KwIf:
            leading := p.pending
            p.pending = nil
            ipos := p.cur.Pos
            p.next()
            // optional parens around condition
            var cond ast.Expr
            if p.cur.Kind == token.LParenSym {
                p.next()
                e, ok := p.parseExprPrec(1)
                if ok {
                    cond = e
                }
                if p.cur.Kind == token.RParenSym {
                    p.next()
                }
            } else {
                e, ok := p.parseExprPrec(1)
                if ok {
                    cond = e
                }
            }
            // then block
            if p.cur.Kind != token.LBraceSym {
                p.errf("expected '{' to start if-block, got %q", p.cur.Lexeme)
                p.syncUntil(token.RBraceSym, token.SemiSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            tblk, err := p.parseFuncBlock()
            if err != nil {
                p.errf("%v", err)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            // optional else block
            var eblk *ast.BlockStmt
            if p.cur.Kind == token.KwElse {
                p.next()
                if p.cur.Kind != token.LBraceSym {
                    p.errf("expected '{' after else, got %q", p.cur.Lexeme)
                } else {
                    if b, err := p.parseFuncBlock(); err == nil {
                        eblk = b
                    } else {
                        p.errf("%v", err)
                    }
                }
            }
            is := &ast.IfStmt{Pos: ipos, Leading: leading, Cond: cond, Then: tblk, Else: eblk}
            stmts = append(stmts, is)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        case token.KwReturn:
            pos := p.cur.Pos
            p.next()
            var results []ast.Expr
            if p.isExprStart(p.cur.Kind) {
                e, ok := p.parseExpr()
                if ok {
                    results = append(results, e)
                }
                for p.cur.Kind == token.CommaSym {
                    p.next()
                    e, ok = p.parseExpr()
                    if ok {
                        results = append(results, e)
                    }
                }
            }
            rs := &ast.ReturnStmt{Pos: pos, Results: results, Leading: p.pending}
            p.pending = nil
            stmts = append(stmts, rs)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        case token.KwVar:
            leading := p.pending
            p.pending = nil
            pos := p.cur.Pos
            p.next()
            if p.cur.Kind != token.Ident {
                p.errf("expected var name, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            namePos := p.cur.Pos
            name := p.cur.Lexeme
            p.next()
            var tname string
            var tpos source.Position
            if p.isTypeName(p.cur.Kind) {
                tname = p.cur.Lexeme
                tpos = p.cur.Pos
                p.next()
                // Allow qualified type like pkg.Type in var declarations
                tname = p.captureQualifiedType(tname)
                // Capture Struct{...} var types fully
                if tname == "Struct" && p.cur.Kind == token.LBraceSym {
                    startOff := tpos.Offset
                    depthBrace := 0
                    depthAngle := 0
                    var lastBraceOff int
                    for {
                        if p.cur.Kind == token.EOF {
                            break
                        }
                        switch p.cur.Kind {
                        case token.LBraceSym:
                            depthBrace++
                        case token.RBraceSym:
                            if depthBrace > 0 {
                                depthBrace--
                            }
                            lastBraceOff = p.cur.Pos.Offset
                            p.next()
                            if depthBrace == 0 {
                                goto doneVarStruct3
                            }
                            continue
                        case token.Lt:
                            depthAngle++
                        case token.Shl:
                            depthAngle += 2
                        case token.Gt:
                            if depthAngle > 0 {
                                depthAngle--
                            }
                        case token.Shr:
                            if depthAngle >= 2 {
                                depthAngle -= 2
                            }
                        }
                        p.next()
                    }
                doneVarStruct3:
                    src := p.s.FileContent()
                    if lastBraceOff > startOff && lastBraceOff+1 <= len(src) {
                        tname = src[startOff : lastBraceOff+1]
                    }
                }
                // Capture optional generic arguments like '<T[,U]>' preserving nested forms
                if p.cur.Kind == token.Lt || p.cur.Kind == token.Shl {
                    startOff := tpos.Offset
                    depth := 0
                    var lastGtOff int
                    for {
                        if p.cur.Kind == token.EOF {
                            break
                        }
                        switch p.cur.Kind {
                        case token.Lt:
                            depth++
                        case token.Shl:
                            depth += 2
                        case token.Gt:
                            depth--
                            lastGtOff = p.cur.Pos.Offset
                            p.next()
                            if depth == 0 {
                                goto doneVarGeneric
                            }
                            continue
                        case token.Shr:
                            depth -= 2
                            lastGtOff = p.cur.Pos.Offset + 1
                            p.next()
                            if depth == 0 {
                                goto doneVarGeneric
                            }
                            continue
                        }
                        p.next()
                    }
                doneVarGeneric:
                    src := p.s.FileContent()
                    if lastGtOff > startOff && lastGtOff+1 <= len(src) {
                        tname = src[startOff : lastGtOff+1]
                    }
                }
            }
            var init ast.Expr
            if p.cur.Kind == token.Assign {
                p.next()
                e, ok := p.parseExpr()
                if ok {
                    init = e
                }
            }
            vd := &ast.VarDecl{Pos: pos, Name: name, NamePos: namePos, Type: tname, TypePos: tpos, Init: init, Leading: leading}
            stmts = append(stmts, vd)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        case token.KwDefer:
            leading := p.pending
            p.pending = nil
            dpos := p.cur.Pos
            p.next()
            callExpr, ok := p.parseExpr()
            if !ok {
                p.errf("expected call after defer, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            ce, _ := callExpr.(*ast.CallExpr)
            if ce == nil {
                p.errf("defer requires a call expression")
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            ds := &ast.DeferStmt{Pos: dpos, Call: ce, Leading: leading}
            stmts = append(stmts, ds)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        case token.Star:
            // Mutating assignment: *name = expr
            leading := p.pending
            p.pending = nil
            starPos := p.cur.Pos
            p.next()
            if p.cur.Kind != token.Ident {
                p.errf("expected identifier after '*' in assignment, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            name := p.cur.Lexeme
            npos := p.cur.Pos
            p.next()
            if p.cur.Kind != token.Assign {
                p.errf("expected '=' in mutating assignment, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            p.next()
            rhs, ok := p.parseExpr()
            if !ok {
                p.errf("expected expression on right-hand side, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            ms := &ast.AssignStmt{Pos: starPos, Name: name, NamePos: npos, Value: rhs, Leading: leading, Mutating: true, StarPos: starPos}
            stmts = append(stmts, ms)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        case token.Ident, token.String, token.Number, token.DurationLit, token.KwSlice, token.KwSet, token.KwMap, token.LParenSym,
            token.Bang, token.Minus, token.TildeSym:
            leading := p.pending
            p.pending = nil
            lhs, ok := p.parseExpr()
            if !ok {
                p.errf("expected expression, got %q", p.cur.Lexeme)
                p.syncUntil(token.SemiSym, token.RBraceSym)
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            // optional assignment operator: only for name = expr
            if p.cur.Kind == token.Assign {
                // ensure lhs is an identifier
                if id, ok := lhs.(*ast.IdentExpr); ok {
                    p.next()
                    rhs, ok := p.parseExpr()
                    if !ok {
                        p.errf("expected expression on right-hand side, got %q", p.cur.Lexeme)
                        p.syncUntil(token.SemiSym, token.RBraceSym)
                        if p.cur.Kind == token.SemiSym {
                            p.next()
                        }
                        continue
                    }
                    as := &ast.AssignStmt{Pos: id.Pos, Name: id.Name, NamePos: id.Pos, Value: rhs, Leading: leading}
                    stmts = append(stmts, as)
                    if p.cur.Kind == token.SemiSym {
                        p.next()
                    }
                    continue
                }
            }
            // otherwise, expression statement
            es := &ast.ExprStmt{Pos: ePos(lhs), X: lhs, Leading: leading}
            stmts = append(stmts, es)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        default:
            // tolerate unknown tokens; attempt to resync at semicolon or '}'
            p.errf("unexpected token in block: %q", p.cur.Lexeme)
            p.syncUntil(token.SemiSym, token.RBraceSym)
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
        }
    }
    if p.cur.Kind != token.RBraceSym {
        return nil, fmt.Errorf("expected '}', got %q", p.cur.Lexeme)
    }
    rb := p.cur.Pos
    p.next()
    return &ast.BlockStmt{LBrace: lb, RBrace: rb, Stmts: stmts}, nil
}
