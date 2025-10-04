package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parseStepBlock parses a block of pipeline-like step statements into a BlockStmt
// with its Stmts populated by StepStmt nodes, tolerantly until the matching '}'.
func (p *Parser) parseStepBlock() (*ast.BlockStmt, error) {
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
        if p.cur.Kind == token.Ident || p.cur.Kind == token.KwIngress || p.cur.Kind == token.KwEgress ||
            p.cur.Kind == token.KwNodeTransform || p.cur.Kind == token.KwNodeFanout || p.cur.Kind == token.KwNodeCollect || p.cur.Kind == token.KwNodeMutable {
            // helper to parse a step body (args + attrs) given a consumed name
            parseStepBody := func(startName string, startPos source.Position) *ast.StepStmt {
                st := &ast.StepStmt{Pos: startPos, Name: startName, Leading: p.pending}
                p.pending = nil
                if p.cur.Kind == token.LParenSym {
                    st.LParen = p.cur.Pos
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
                        if p.cur.Kind == token.CommaSym {
                            p.next()
                            continue
                        }
                    }
                    if p.cur.Kind == token.RParenSym {
                        st.RParen = p.cur.Pos
                        p.next()
                    } else {
                        p.errf("missing ')' in step args")
                    }
                    st.Args = args
                }
                var attrs []ast.Attr
                for p.cur.Kind == token.Ident || p.cur.Kind == token.KwType {
                    aname := p.cur.Lexeme
                    apos := p.cur.Pos
                    p.next()
                    for p.cur.Kind == token.DotSym {
                        p.next()
                        if !(p.cur.Kind == token.Ident || p.cur.Kind == token.KwType || p.cur.Kind == token.KwPipeline) {
                            p.errf("expected ident after '.' in attribute name, got %q", p.cur.Lexeme)
                            break
                        }
                        aname += "." + p.cur.Lexeme
                        p.next()
                    }
                    var aargs []ast.Arg
                    if p.cur.Kind == token.LParenSym {
                        p.next()
                        for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                            if arg, ok := p.parseAttrArg(); ok {
                                aargs = append(aargs, arg)
                            } else {
                                p.errf("unexpected token in attr args: %q", p.cur.Lexeme)
                                p.syncUntil(token.CommaSym, token.RParenSym)
                            }
                            if p.cur.Kind == token.CommaSym {
                                p.next()
                                continue
                            }
                        }
                        if p.cur.Kind == token.RParenSym {
                            p.next()
                        } else {
                            p.errf("missing ')' in attr call")
                        }
                    }
                    attrs = append(attrs, ast.Attr{Pos: apos, Name: aname, Args: aargs})
                    if p.cur.Kind == token.CommaSym {
                        p.next()
                        continue
                    }
                }
                st.Attrs = attrs
                return st
            }

            // parse a step name
            name := p.cur.Lexeme
            namePos := p.cur.Pos
            p.next()
            st := parseStepBody(name, namePos)
            stmts = append(stmts, st)
            // chained steps: .Name(args) ...
            for p.cur.Kind == token.DotSym {
                p.next()
                // expect next step name
                if !(p.cur.Kind == token.Ident || p.cur.Kind == token.KwIngress || p.cur.Kind == token.KwEgress || p.cur.Kind == token.KwNodeTransform || p.cur.Kind == token.KwNodeFanout || p.cur.Kind == token.KwNodeCollect || p.cur.Kind == token.KwNodeMutable) {
                    p.errf("expected node name after '.', got %q", p.cur.Lexeme)
                    break
                }
                cname := p.cur.Lexeme
                cpos := p.cur.Pos
                p.next()
                // support dotted step names like pkg.Func after chain dot
                for p.cur.Kind == token.DotSym {
                    p.next()
                    if p.cur.Kind != token.Ident {
                        p.errf("expected identifier after '.', got %q", p.cur.Lexeme)
                        break
                    }
                    cname = cname + "." + p.cur.Lexeme
                    p.next()
                }
                cst := parseStepBody(cname, cpos)
                stmts = append(stmts, cst)
            }
            if p.cur.Kind == token.SemiSym {
                p.next()
            }
            continue
        }
        // unknown tokens: recover to semicolon or '}'
        p.errf("unexpected token in error block: %q", p.cur.Lexeme)
        p.syncUntil(token.SemiSym, token.RBraceSym)
        if p.cur.Kind == token.SemiSym {
            p.next()
        }
    }
    rb := p.cur.Pos
    if p.cur.Kind == token.RBraceSym {
        p.next()
    }
    return &ast.BlockStmt{LBrace: lb, RBrace: rb, Stmts: stmts}, nil
}
