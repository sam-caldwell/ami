package parser

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// parsePipelineDecl parses: pipeline IDENT { Node(args) ('.'|'->') Node(args) ... } [ error { NodeChain } ]
func (p *Parser) parsePipelineDecl() astpkg.PipelineDecl {
    // consume 'pipeline'
    p.next()
    name := ""
    if p.cur.Kind == tok.IDENT {
        name = p.cur.Lexeme
        p.next()
    }
    // require '{'
    if p.cur.Kind != tok.LBRACE {
        return astpkg.PipelineDecl{}
    }
    p.next()
    steps, connectors := p.parseNodeChain()
    // require '}'
    if p.cur.Kind == tok.RBRACE {
        p.next()
    }
    // optional error pipeline: 'error' '{' NodeChain '}'
    errSteps := []astpkg.NodeCall{}
    errConns := []string{}
    if p.cur.Kind == tok.KW_ERROR || (p.cur.Kind == tok.IDENT && strings.ToLower(p.cur.Lexeme) == "error") {
        p.next()
        if p.cur.Kind == tok.LBRACE {
            p.next()
            errSteps, errConns = p.parseNodeChain()
            if p.cur.Kind == tok.RBRACE {
                p.next()
            }
        }
    }
    return astpkg.PipelineDecl{Name: name, Steps: steps, Connectors: connectors, ErrorSteps: errSteps, ErrorConnectors: errConns}
}

// parseNodeChain parses Node(args) ('.'|'->') Node(args) ... until '}' or EOF
func (p *Parser) parseNodeChain() ([]astpkg.NodeCall, []string) {
    steps := []astpkg.NodeCall{}
    connectors := []string{}
    // Expect first node
    n, ok := p.parseNodeCall()
    if !ok {
        return steps, connectors
    }
    steps = append(steps, n)
    // Zero or more (. or ->) NodeCall
    for p.cur.Kind == tok.DOT || p.cur.Kind == tok.ARROW {
        conn := "."
        if p.cur.Kind == tok.ARROW {
            conn = "->"
        }
        p.next()
        n2, ok := p.parseNodeCall()
        if !ok {
            break
        }
        connectors = append(connectors, conn)
        steps = append(steps, n2)
    }
    return steps, connectors
}

// parseNodeCall parses IDENT '(' args ')'
func (p *Parser) parseNodeCall() (astpkg.NodeCall, bool) {
    if !(p.cur.Kind == tok.IDENT || p.cur.Kind == tok.KW_INGRESS || p.cur.Kind == tok.KW_TRANSFORM || p.cur.Kind == tok.KW_FANOUT || p.cur.Kind == tok.KW_COLLECT || p.cur.Kind == tok.KW_EGRESS) {
        return astpkg.NodeCall{}, false
    }
    pending := p.consumeComments()
    startTok := p.cur
    name := p.cur.Lexeme
    if name == "" { // for keyword tokens, Lexeme may carry source; fall back to kind name
        switch p.cur.Kind {
        case tok.KW_INGRESS:
            name = "ingress"
        case tok.KW_TRANSFORM:
            name = "transform"
        case tok.KW_FANOUT:
            name = "fanout"
        case tok.KW_COLLECT:
            name = "collect"
        case tok.KW_EGRESS:
            name = "egress"
        }
    }
    p.next()
    if p.cur.Kind != tok.LPAREN {
        return astpkg.NodeCall{Name: name}, true
    }
    // parse arguments as raw strings; handle nesting and commas at depth=1
    p.next() // consume '('
    args := []string{}
    workers := []astpkg.WorkerRef{}
    var buf strings.Builder
    depth := 1
    for p.cur.Kind != tok.EOF && depth > 0 {
        switch p.cur.Kind {
        case tok.LPAREN:
            depth++
            buf.WriteString("(")
            p.next()
        case tok.RPAREN:
            depth--
            if depth == 0 { // finish current arg if non-empty
                s := strings.TrimSpace(buf.String())
                if s != "" {
                    args = append(args, s)
                    if w, ok := parseWorkerRef(s); ok {
                        workers = append(workers, w)
                    }
                }
                buf.Reset()
                p.next()
                break
            }
            buf.WriteString(")")
            p.next()
        case tok.COMMA:
            if depth == 1 {
                s := strings.TrimSpace(buf.String())
                args = append(args, s)
                if w, ok := parseWorkerRef(s); ok {
                    workers = append(workers, w)
                }
                buf.Reset()
                p.next()
                continue
            }
            buf.WriteString(",")
            p.next()
        default:
            if p.cur.Lexeme != "" {
                buf.WriteString(p.cur.Lexeme)
            } else {
                buf.WriteRune(rune(p.cur.Kind))
            }
            p.next()
        }
    }
    return astpkg.NodeCall{Name: name, Args: args, Workers: workers, Pos: p.posFrom(startTok), Comments: pending}, true
}

