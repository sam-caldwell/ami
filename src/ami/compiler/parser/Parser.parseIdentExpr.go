package parser

import (
    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseIdentExpr parses an identifier-led expression where the initial ident
// has already been consumed. It supports dotted selector chains, container
// literals, and calls.
func (p *Parser) parseIdentExpr(first string, firstPos source.Position) ast.Expr {
    // gather dotted name parts
    parts := []string{first}
    poses := []source.Position{firstPos}
    for p.cur.Kind == token.DotSym {
        p.next()
        if p.cur.Kind != token.Ident {
            p.errf("expected ident after '.', got %q", p.cur.Lexeme)
            break
        }
        parts = append(parts, p.cur.Lexeme)
        poses = append(poses, p.cur.Pos)
        p.next()
    }
    base := parts[0]
    // container literals after a bare keyword-like name
    if p.cur.Kind == token.Lt {
        switch base {
        case "slice":
            if lit, ok := p.parseSliceOrSetLiteral(true, firstPos); ok {
                return lit
            }
        case "set":
            if lit, ok := p.parseSliceOrSetLiteral(false, firstPos); ok {
                return lit
            }
        case "map":
            if lit, ok := p.parseMapLiteral(firstPos); ok {
                return lit
            }
        }
    }
    if p.cur.Kind == token.LParenSym {
        // join dotted parts for call names
        full := parts[0]
        for i := 1; i < len(parts); i++ {
            full += "." + parts[i]
        }
        lp := p.cur.Pos
        p.next()
        var args []ast.Expr
        for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
            e, ok := p.parseExprPrec(1)
            if ok {
                // Named arg detection: ident followed by '='
                if p.cur.Kind == token.Assign {
                    if _, isIdent := e.(*ast.IdentExpr); isIdent {
                        // consume '=' and parse value
                        p.next()
                        if p.cur.Kind == token.LBracketSym {
                            // consume balanced brackets
                            depth := 0
                            for {
                                if p.cur.Kind == token.EOF { break }
                                if p.cur.Kind == token.LBracketSym { depth++ }
                                if p.cur.Kind == token.RBracketSym {
                                    if depth > 0 { depth-- }
                                    p.next()
                                    if depth == 0 { break }
                                    continue
                                }
                                p.next()
                            }
                            args = append(args, &ast.IdentExpr{Pos: ePos(e), Name: "_"})
                        } else if p.isAttrBareValueKeyword(p.cur.Kind) {
                            p.next()
                            args = append(args, &ast.IdentExpr{Pos: ePos(e), Name: "_"})
                        } else if p.cur.Kind == token.KwEvent {
                            name := p.cur.Lexeme
                            pos := p.cur.Pos
                            p.next()
                            _ = p.parseIdentExpr(name, pos)
                            args = append(args, &ast.IdentExpr{Pos: ePos(e), Name: "_"})
                        } else {
                            if v, ok2 := p.parseExprPrec(1); ok2 {
                                _ = v
                                args = append(args, &ast.IdentExpr{Pos: ePos(e), Name: "_"})
                            } else {
                                p.errf("unexpected token in call args: %q", p.cur.Lexeme)
                                p.syncUntil(token.CommaSym, token.RParenSym)
                            }
                        }
                    } else {
                        // malformed named arg; treat as positional
                        args = append(args, e)
                    }
                } else {
                    // positional arg
                    args = append(args, e)
                }
            } else {
                p.errf("unexpected token in call args: %q", p.cur.Lexeme)
                p.syncUntil(token.CommaSym, token.RParenSym)
            }
            if p.cur.Kind == token.CommaSym { p.next(); continue }
        }
        rp := p.cur.Pos
        if p.cur.Kind == token.RParenSym {
            p.next()
        } else {
            p.errf("missing ')' in call expr")
        }
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
