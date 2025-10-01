package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

func (p *Parser) parsePipelineDecl() (*ast.PipelineDecl, error) {
    pos := p.cur.Pos
    p.next()
    if p.cur.Kind != token.Ident {
        return nil, fmt.Errorf("expected pipeline name, got %q", p.cur.Lexeme)
    }
    name := p.cur.Lexeme
    namePos := p.cur.Pos
    p.next()
    // consume empty param list for now
    if p.cur.Kind != token.LParenSym {
        return nil, fmt.Errorf("expected '(', got %q", p.cur.Lexeme)
    }
    lp := p.cur.Pos
    p.next()
    if p.cur.Kind != token.RParenSym {
        return nil, fmt.Errorf("expected ')', got %q", p.cur.Lexeme)
    }
    rp := p.cur.Pos
    p.next()
    // parse body with optional leading error block and simple steps
    if p.cur.Kind != token.LBraceSym {
        return nil, fmt.Errorf("expected '{', got %q", p.cur.Lexeme)
    }
    lb := p.cur.Pos
    p.next()
    var errblk *ast.ErrorBlock
    var stmts []ast.Stmt
    if p.cur.Kind == token.KwError {
        eb, err := p.parseErrorBlock()
        if err != nil {
            return nil, err
        }
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
                if p.cur.Kind != token.Ident {
                    p.errf("expected identifier after '.', got %q", p.cur.Lexeme)
                    break
                }
                name = name + "." + p.cur.Lexeme
                p.next()
            }
            if p.cur.Kind == token.Arrow {
                // Edge: name -> ident
                p.next()
                if !(p.cur.Kind == token.Ident || p.cur.Kind == token.KwNodeTransform || p.cur.Kind == token.KwNodeFanout || p.cur.Kind == token.KwNodeCollect || p.cur.Kind == token.KwNodeMutable || p.cur.Kind == token.KwIngress || p.cur.Kind == token.KwEgress) {
                    p.errf("expected identifier after '->', got %q", p.cur.Lexeme)
                    p.syncUntil(token.SemiSym, token.RBraceSym)
                    if p.cur.Kind == token.SemiSym {
                        p.next()
                    }
                    continue
                }
                to := p.cur.Lexeme
                toPos := p.cur.Pos
                edge := &ast.EdgeStmt{Pos: namePos, From: name, FromPos: namePos, To: to, ToPos: toPos, Leading: p.pending}
                p.pending = nil
                stmts = append(stmts, edge)
                p.next()
                if p.cur.Kind == token.SemiSym {
                    p.next()
                }
                continue
            }
            // Step call starting with name already consumed
            parseStepBody := func(startName string, startPos source.Position) *ast.StepStmt {
                st := &ast.StepStmt{Pos: startPos, Name: startName, Leading: p.pending}
                p.pending = nil
                // optional args
                if p.cur.Kind == token.LParenSym {
                    p.next()
                    var args []ast.Arg
                    for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                        if arg, ok := p.parseAttrArg(); ok {
                            args = append(args, arg)
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
                        p.next()
                    } else {
                        p.errf("missing ')' in step args")
                    }
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

            // first step
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
        // unknown: recover to semicolon or '}'
        p.errf("unexpected token in pipeline: %q", p.cur.Lexeme)
        p.syncUntil(token.SemiSym, token.RBraceSym)
        if p.cur.Kind == token.SemiSym {
            p.next()
        }
    }
    if p.cur.Kind != token.RBraceSym {
        return nil, fmt.Errorf("expected '}', got %q", p.cur.Lexeme)
    }
    rb := p.cur.Pos
    p.next()
    pd := &ast.PipelineDecl{Pos: pos, Name: name, NamePos: namePos, Body: &ast.BlockStmt{LBrace: lb, RBrace: rb}, Error: errblk, Leading: p.pending, Stmts: stmts, LParen: lp, RParen: rp}
    p.pending = nil
    return pd, nil
}

