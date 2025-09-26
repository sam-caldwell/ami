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
    attrs := map[string]string{}
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
    // Derive structured attributes from top-level args while preserving args for compatibility
    // Recognized keys
    isKey := func(k string) bool {
        switch k {
        case "in", "worker", "minWorkers", "maxWorkers", "onError", "capabilities", "type":
            return true
        default:
            return false
        }
    }
    for _, a := range args {
        // split on first '=' only
        if eq := strings.IndexByte(a, '='); eq > 0 {
            key := strings.TrimSpace(a[:eq])
            val := strings.TrimSpace(a[eq+1:])
                if isKey(key) {
                attrs[key] = val
                if key == "worker" {
                    // Inline literal: do not treat as a reference worker
                    if strings.HasPrefix(val, "func") {
                        // Tokenize the value and parse a primary expression
                        p2 := New(val)
                        // collect tokens until EOF
                        var toks []tok.Token
                        for p2.cur.Kind != tok.EOF {
                            toks = append(toks, p2.cur)
                            p2.next()
                        }
                        bp := &bodyParser{toks: toks}
                        if expr, ok := bp.parsePrimary(); ok {
                            if fl, ok2 := expr.(astpkg.FuncLit); ok2 {
                                // allocate to set pointer
                                lit := fl
                                if attrs == nil { attrs = map[string]string{} }
                                // store on NodeCall below
                                // We'll attach after loop using local variable
                                _ = lit // placeholder for clarity
                            }
                        }
                    } else {
                        if w, ok := parseWorkerRef(val); ok {
                            workers = append(workers, w)
                        }
                    }
                }
            }
        }
    }
    // Attach inline worker literal if parsed
    var inline *astpkg.FuncLit
    if wsrc, ok := attrs["worker"]; ok && strings.HasPrefix(wsrc, "func") {
        p2 := New(wsrc)
        var toks []tok.Token
        for p2.cur.Kind != tok.EOF { toks = append(toks, p2.cur); p2.next() }
        bp := &bodyParser{toks: toks}
        if expr, ok := bp.parsePrimary(); ok {
            if fl, ok2 := expr.(astpkg.FuncLit); ok2 {
                lit := fl
                inline = &lit
            }
        }
    }
    return astpkg.NodeCall{Name: name, Args: args, Attrs: attrs, InlineWorker: inline, Workers: workers, Pos: p.posFrom(startTok), Comments: pending}, true
}
