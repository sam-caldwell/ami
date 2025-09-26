package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// bodyParser parses simple statements from captured tokens.
type bodyParser struct {
    toks     []tok.Token
    i        int
    comments map[int][]astpkg.Comment
}

func (bp *bodyParser) atEnd() bool { return bp.i >= len(bp.toks) }
func (bp *bodyParser) cur() tok.Token {
    if bp.atEnd() {
        return tok.Token{Kind: tok.EOF}
    }
    return bp.toks[bp.i]
}
func (bp *bodyParser) next() {
    if !bp.atEnd() {
        bp.i++
    }
}

func parseBodyStmts(toks []tok.Token, commentMap map[int][]astpkg.Comment) []astpkg.Stmt {
    bp := &bodyParser{toks: toks, comments: commentMap}
    var out []astpkg.Stmt
    for !bp.atEnd() {
        if s, ok := bp.parseStmt(); ok {
            out = append(out, s)
            continue
        }
        bp.next()
    }
    return out
}

func (bp *bodyParser) parseStmt() (astpkg.Stmt, bool) {
    toPos := func(t tok.Token) astpkg.Position { return astpkg.Position{Line: t.Line, Column: t.Column, Offset: t.Offset} }
    if bp.cur().Kind == tok.KW_VAR {
        start := bp.cur()
        bp.next()
        if bp.cur().Kind != tok.IDENT {
            return nil, false
        }
        nameTok := bp.cur()
        name := nameTok.Lexeme
        bp.next()
        tr, hasType := bp.parseTypeRef()
        var init astpkg.Expr
        if bp.cur().Kind == tok.ASSIGN {
            bp.next()
            if x, ok := bp.parseExpr(); ok { init = x }
        }
        if hasType {
            return bp.attachStmtComments(start, astpkg.VarDeclStmt{Name: name, Type: tr, Init: init, Pos: toPos(start)}), true
        }
        return bp.attachStmtComments(start, astpkg.VarDeclStmt{Name: name, Init: init, Pos: toPos(start)}), true
    }
    if bp.cur().Kind == tok.KW_DEFER {
        start := bp.cur()
        bp.next()
        if x, ok := bp.parseExpr(); ok {
            return bp.attachStmtComments(start, astpkg.DeferStmt{X: x, Pos: toPos(start)}), true
        }
        return bp.attachStmtComments(start, astpkg.DeferStmt{X: nil, Pos: toPos(start)}), true
    }
    if bp.cur().Kind == tok.KW_RETURN {
        start := bp.cur()
        bp.next()
        var results []astpkg.Expr
        if x, ok := bp.parseExpr(); ok {
            results = append(results, x)
            for bp.cur().Kind == tok.COMMA {
                bp.next()
                if y, ok2 := bp.parseExpr(); ok2 { results = append(results, y) } else { break }
            }
        }
        return bp.attachStmtComments(start, astpkg.ReturnStmt{Results: results, Pos: toPos(start)}), true
    }
    // assignment or call expr
    save := bp.i
    if lhs, ok := bp.parseExpr(); ok {
        if bp.cur().Kind == tok.ASSIGN {
            assignTok := bp.cur()
            bp.next()
            if rhs, ok2 := bp.parseExpr(); ok2 {
                return bp.attachStmtComments(bp.toks[save], astpkg.AssignStmt{LHS: lhs, RHS: rhs, Pos: toPos(assignTok)}), true
            }
            bp.i = save
        } else {
            if start := save; start < len(bp.toks) {
                return bp.attachStmtComments(bp.toks[start], astpkg.ExprStmt{X: lhs, Pos: toPos(bp.toks[start])}), true
            }
            return bp.attachStmtComments(tok.Token{}, astpkg.ExprStmt{X: lhs}), true
        }
    }
    bp.i = save
    return nil, false
}

func (bp *bodyParser) attachStmtComments(start tok.Token, s astpkg.Stmt) astpkg.Stmt {
    var off int
    if start.Kind != tok.EOF { off = start.Offset } else if bp.i < len(bp.toks) { off = bp.toks[bp.i].Offset }
    if off != 0 {
        if cmts, ok := bp.comments[off]; ok {
            switch v := s.(type) {
            case astpkg.AssignStmt: v.Comments = append(v.Comments, cmts...); return v
            case astpkg.ExprStmt: v.Comments = append(v.Comments, cmts...); return v
            case astpkg.DeferStmt: v.Comments = append(v.Comments, cmts...); return v
            case astpkg.ReturnStmt: v.Comments = append(v.Comments, cmts...); return v
            case astpkg.VarDeclStmt: v.Comments = append(v.Comments, cmts...); return v
            case astpkg.BlockStmt: v.Comments = append(v.Comments, cmts...); return v
            case astpkg.MutBlockStmt: v.Comments = append(v.Comments, cmts...); return v
            }
        }
    }
    return s
}

// --- Binary expression parsing (precedence-climbing) ---

func (bp *bodyParser) parseExpr() (astpkg.Expr, bool) { return bp.parseBinaryExpr() }

func (bp *bodyParser) parseBinaryExpr() (astpkg.Expr, bool) {
    left, ok := bp.parsePrimary()
    if !ok { return nil, false }
    return bp.parseBinaryRHS(0, left)
}

func (bp *bodyParser) precedence(k tok.Kind) int {
    switch k {
    case tok.STAR, tok.SLASH, tok.PERCENT: return 40
    case tok.PLUS, tok.MINUS: return 30
    case tok.LT, tok.LTE, tok.GT, tok.GTE: return 20
    case tok.EQ, tok.NEQ: return 10
    default: return -1
    }
}

func (bp *bodyParser) parseBinaryRHS(minPrec int, left astpkg.Expr) (astpkg.Expr, bool) {
    for {
        beforeI := bp.i
        opTok := bp.cur()
        prec := bp.precedence(opTok.Kind)
        if prec < minPrec { break }
        // treat '*' as multiplication only when not the LHS mutation marker: pattern '*' IDENT '='
        if opTok.Kind == tok.STAR {
            if bp.i+2 < len(bp.toks) && bp.toks[bp.i+1].Kind == tok.IDENT && bp.toks[bp.i+2].Kind == tok.ASSIGN { break }
        }
        bp.next()
        right, ok := bp.parsePrimary()
        if !ok { return left, true }
        for {
            nextPrec := bp.precedence(bp.cur().Kind)
            // Do not cross a mutating assignment boundary: '* IDENT ='
            if bp.cur().Kind == tok.STAR {
                if bp.i+2 < len(bp.toks) && bp.toks[bp.i+1].Kind == tok.IDENT && bp.toks[bp.i+2].Kind == tok.ASSIGN {
                    nextPrec = -1
                }
            }
            if nextPrec > prec {
                var ok2 bool
                right, ok2 = bp.parseBinaryRHS(prec+1, right)
                if !ok2 { break }
            } else { break }
        }
        op := opTok.Lexeme
        if op == "" {
            switch opTok.Kind {
            case tok.STAR: op = "*"
            case tok.SLASH: op = "/"
            case tok.PERCENT: op = "%"
            case tok.PLUS: op = "+"
            case tok.MINUS: op = "-"
            case tok.LT: op = "<"
            case tok.LTE: op = "<="
            case tok.GT: op = ">"
            case tok.GTE: op = ">="
            case tok.EQ: op = "=="
            case tok.NEQ: op = "!="
            }
        }
        left = astpkg.BinaryExpr{X: left, Op: op, Y: right}
        if bp.i == beforeI { break }
    }
    return left, true
}

func (bp *bodyParser) parsePrimary() (astpkg.Expr, bool) {
    t := bp.cur()
    toPos := func(tk tok.Token) astpkg.Position { return astpkg.Position{Line: tk.Line, Column: tk.Column, Offset: tk.Offset} }
    switch t.Kind {
    case tok.KW_SLICE, tok.KW_SET:
        kind := "slice"; if t.Kind == tok.KW_SET { kind = "set" }
        start := t
        bp.next()
        var args []astpkg.TypeRef
        if bp.cur().Kind == tok.LT { bp.next(); if a, ok := bp.parseTypeRef(); ok { args = append(args, a) }; if bp.cur().Kind == tok.GT { bp.next() } }
        if bp.cur().Kind != tok.LBRACE { return nil, false }
        bp.next()
        var elems []astpkg.Expr
        for !bp.atEnd() && bp.cur().Kind != tok.RBRACE {
            if bp.cur().Kind == tok.COMMA { bp.next(); continue }
            if e, ok := bp.parseExpr(); ok { elems = append(elems, e) } else { bp.next() }
            if bp.cur().Kind == tok.COMMA { bp.next() }
        }
        if bp.cur().Kind == tok.RBRACE { bp.next() }
        return astpkg.ContainerLit{Kind: kind, TypeArgs: args, Elems: elems, Pos: toPos(start)}, true
    case tok.KW_MAP:
        start := t
        bp.next()
        var args []astpkg.TypeRef
        if bp.cur().Kind == tok.LT {
            bp.next()
            if a, ok := bp.parseTypeRef(); ok { args = append(args, a) }
            if bp.cur().Kind == tok.COMMA { bp.next() }
            if b, ok := bp.parseTypeRef(); ok { args = append(args, b) }
            if bp.cur().Kind == tok.GT { bp.next() }
        }
        if bp.cur().Kind != tok.LBRACE { return nil, false }
        bp.next()
        var kvs []astpkg.MapElem
        for !bp.atEnd() && bp.cur().Kind != tok.RBRACE {
            if bp.cur().Kind == tok.COMMA { bp.next(); continue }
            k, ok := bp.parseExpr(); if !ok { bp.next(); continue }
            if bp.cur().Kind != tok.COLON { bp.next(); continue }
            bp.next()
            v, ok2 := bp.parseExpr(); if !ok2 { bp.next(); continue }
            kvs = append(kvs, astpkg.MapElem{Key: k, Value: v})
            if bp.cur().Kind == tok.COMMA { bp.next() }
        }
        if bp.cur().Kind == tok.RBRACE { bp.next() }
        return astpkg.ContainerLit{Kind: "map", TypeArgs: args, MapElems: kvs, Pos: toPos(start)}, true
    case tok.KW_STATE:
        // Treat reserved 'state' as an identifier for method calls/selectors
        name := "state"
        bp.next()
        if bp.cur().Kind == tok.DOT {
            recv := astpkg.Ident{Name: name, Pos: toPos(t)}
            bp.next()
            if bp.cur().Kind == tok.IDENT {
                sel := bp.cur().Lexeme
                selTok := bp.cur()
                bp.next()
                if bp.cur().Kind == tok.LPAREN { args := bp.parseArgs(); return astpkg.CallExpr{Fun: astpkg.SelectorExpr{X: recv, Sel: sel, Pos: toPos(selTok)}, Args: args, Pos: toPos(t)}, true }
                return astpkg.SelectorExpr{X: recv, Sel: sel, Pos: toPos(selTok)}, true
            }
            return astpkg.Ident{Name: name, Pos: toPos(t)}, true
        }
        if bp.cur().Kind == tok.LPAREN { args := bp.parseArgs(); return astpkg.CallExpr{Fun: astpkg.Ident{Name: name, Pos: toPos(t)}, Args: args, Pos: toPos(t)}, true }
        return astpkg.Ident{Name: name, Pos: toPos(t)}, true
    case tok.IDENT:
        name := t.Lexeme
        bp.next()
        // Optional type arguments: IDENT '<' Type {',' Type} '>'
        var typeArgs []astpkg.TypeRef
        if bp.cur().Kind == tok.LT {
            bp.next()
            for !bp.atEnd() {
                if bp.cur().Kind == tok.GT { bp.next(); break }
                if bp.cur().Kind == tok.COMMA { bp.next(); continue }
                if a, ok := bp.parseTypeRef(); ok { typeArgs = append(typeArgs, a); continue }
                bp.next()
            }
        }
        if bp.cur().Kind == tok.DOT {
            recv := astpkg.Ident{Name: name, Pos: toPos(t)}
            bp.next()
            if bp.cur().Kind == tok.IDENT {
                sel := bp.cur().Lexeme
                selTok := bp.cur()
                bp.next()
                if bp.cur().Kind == tok.LPAREN { args := bp.parseArgs(); return astpkg.CallExpr{Fun: astpkg.SelectorExpr{X: recv, Sel: sel, Pos: toPos(selTok)}, Args: args, TypeArgs: typeArgs, Pos: toPos(t)}, true }
                return astpkg.SelectorExpr{X: recv, Sel: sel, Pos: toPos(selTok)}, true
            }
            return astpkg.Ident{Name: name, Pos: toPos(t)}, true
        }
        if bp.cur().Kind == tok.LPAREN { args := bp.parseArgs(); return astpkg.CallExpr{Fun: astpkg.Ident{Name: name, Pos: toPos(t)}, Args: args, TypeArgs: typeArgs, Pos: toPos(t)}, true }
        return astpkg.Ident{Name: name, Pos: toPos(t)}, true
    case tok.KW_FUNC:
        // Minimal function literal: capture body tokens; params/results omitted (scaffold)
        start := t
        bp.next()
        // Skip params if present: '(' ... ')'
        if bp.cur().Kind == tok.LPAREN {
            depth := 1
            bp.next()
            for !bp.atEnd() && depth > 0 {
                if bp.cur().Kind == tok.LPAREN { depth++ }
                if bp.cur().Kind == tok.RPAREN { depth--; if depth == 0 { bp.next(); break } }
                bp.next()
            }
        }
        // Optional result: either '(' ... ')' or a single type; skip heuristically
        if bp.cur().Kind == tok.LPAREN {
            depth := 1
            bp.next()
            for !bp.atEnd() && depth > 0 {
                if bp.cur().Kind == tok.LPAREN { depth++ }
                if bp.cur().Kind == tok.RPAREN { depth--; if depth == 0 { bp.next(); break } }
                bp.next()
            }
        } else {
            // best-effort skip a single type token sequence until '{' or '('
            for !bp.atEnd() && bp.cur().Kind != tok.LBRACE && bp.cur().Kind != tok.LPAREN {
                // stop on separators
                if bp.cur().Kind == tok.COMMA || bp.cur().Kind == tok.SEMI { break }
                bp.next()
            }
        }
        // Body tokens between '{' and matching '}'
        var body []tok.Token
        if bp.cur().Kind == tok.LBRACE {
            depth := 1
            bp.next()
            for !bp.atEnd() && depth > 0 {
                if bp.cur().Kind == tok.LBRACE { depth++ }
                if bp.cur().Kind == tok.RBRACE { depth--; if depth == 0 { bp.next(); break } }
                if depth > 0 { body = append(body, bp.cur()) }
                bp.next()
            }
        }
        return astpkg.FuncLit{Body: body, Pos: toPos(start)}, true
    case tok.STRING:
        bp.next(); return astpkg.BasicLit{Kind: "string", Value: t.Lexeme, Pos: toPos(t)}, true
    case tok.NUMBER:
        bp.next(); return astpkg.BasicLit{Kind: "number", Value: t.Lexeme, Pos: toPos(t)}, true
    case tok.STAR:
        bp.next(); if x, ok := bp.parsePrimary(); ok { return astpkg.UnaryExpr{Op: "*", X: x, Pos: toPos(t)}, true }; return nil, false
    case tok.AMP:
        bp.next(); if x, ok := bp.parsePrimary(); ok { return astpkg.UnaryExpr{Op: "&", X: x, Pos: toPos(t)}, true }; return nil, false
    case tok.LPAREN:
        bp.next(); e, ok := bp.parseBinaryExpr(); if bp.cur().Kind == tok.RPAREN { bp.next() }; return e, ok
    default:
        return nil, false
    }
}

func (bp *bodyParser) parseArgs() []astpkg.Expr {
    if bp.cur().Kind != tok.LPAREN { return nil }
    bp.next()
    var out []astpkg.Expr
    depth := 1
    for !bp.atEnd() && depth > 0 {
        switch bp.cur().Kind {
        case tok.LPAREN: depth++; bp.next()
        case tok.RPAREN: depth--; bp.next(); if depth == 0 { break }
        case tok.COMMA: bp.next()
        default:
            if e, ok := bp.parseExpr(); ok { out = append(out, e) } else { bp.next() }
        }
    }
    return out
}

// parseTypeRef mirrors Parser.parseType for local variable declarations.
func (bp *bodyParser) parseTypeRef() (astpkg.TypeRef, bool) {
    var tr astpkg.TypeRef
    if bp.cur().Kind == tok.STAR { bp.next(); return tr, false }
    if bp.cur().Kind == tok.LBRACK { bp.next(); if bp.cur().Kind == tok.RBRACK { tr.Slice = true; bp.next() } else { return tr, false } }
    switch bp.cur().Kind {
    case tok.IDENT, tok.KW_MAP, tok.KW_SET, tok.KW_SLICE:
        tr.Name = bp.cur().Lexeme; bp.next()
    default:
        return tr, false
    }
    if bp.cur().Kind == tok.LT {
        bp.next()
        for !bp.atEnd() {
            if bp.cur().Kind == tok.GT { bp.next(); break }
            if bp.cur().Kind == tok.COMMA { bp.next(); continue }
            if arg, ok := bp.parseTypeRef(); ok { tr.Args = append(tr.Args, arg); continue }
            bp.next()
        }
    }
    return tr, true
}
