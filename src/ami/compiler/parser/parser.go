package parser

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    scan "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

type Parser struct {
    s   *scan.Scanner
    cur tok.Token
    file   string
    errors []diag.Diagnostic
}

func New(src string) *Parser {
    p := &Parser{s: scan.New(src), file: "input.ami"}
    p.next()
    return p
}

// Errors returns collected diagnostics from tolerant parsing.
func (p *Parser) Errors() []diag.Diagnostic { return append([]diag.Diagnostic(nil), p.errors...) }

func (p *Parser) next() { p.cur = p.s.Next() }

func (p *Parser) errorf(msg string) {
    p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PARSE", Message: msg, File: p.file})
}

// synchronize advances the parser until a likely statement boundary is found.
func (p *Parser) synchronize() {
    for p.cur.Kind != tok.EOF {
        if p.cur.Kind == tok.SEMI { p.next(); return }
        switch p.cur.Kind {
        case tok.KW_FUNC, tok.KW_IMPORT, tok.KW_PIPELINE, tok.KW_PACKAGE, tok.RBRACE:
            return
        }
        p.next()
    }
}

func (p *Parser) ParseFile() *astpkg.File {
    f := &astpkg.File{}
    for p.cur.Kind != tok.EOF {
        // package clause
        if p.cur.Kind == tok.KW_PACKAGE || (p.cur.Kind == tok.IDENT && p.cur.Lexeme == "package") {
            p.next()
            if p.cur.Kind == tok.IDENT {
                f.Package = p.cur.Lexeme
                p.next()
                continue
            }
            p.errorf("expected package name")
            p.synchronize()
        }
        // pipeline declaration: pipeline IDENT { <chain> }
        if p.cur.Kind == tok.KW_PIPELINE {
            decl := p.parsePipelineDecl()
            if decl.Name != "" {
                f.Decls = append(f.Decls, decl)
                f.Stmts = append(f.Stmts, decl)
                continue
            }
            p.errorf("invalid pipeline declaration")
            p.synchronize()
            continue
        }
        // import declarations
        if p.cur.Kind == tok.KW_IMPORT || (p.cur.Kind == tok.IDENT && p.cur.Lexeme == "import") {
            p.next()
            // import "path"
            if p.cur.Kind == tok.STRING {
                path := unquote(p.cur.Lexeme)
                f.Imports = append(f.Imports, path)
                f.Decls = append(f.Decls, astpkg.ImportDecl{Path: path})
                f.Stmts = append(f.Stmts, astpkg.ImportDecl{Path: path})
                p.next()
                continue
            }
            // import alias "path" -> skip alias
            if p.cur.Kind == tok.IDENT {
                alias := p.cur.Lexeme
                p.next()
                if p.cur.Kind == tok.STRING {
                    path := unquote(p.cur.Lexeme)
                    f.Imports = append(f.Imports, path)
                    f.Decls = append(f.Decls, astpkg.ImportDecl{Path: path, Alias: alias})
                    f.Stmts = append(f.Stmts, astpkg.ImportDecl{Path: path, Alias: alias})
                    p.next()
                    continue
                }
            }
            // import ( ... )
            if p.cur.Kind == tok.LPAREN {
                p.next()
                for p.cur.Kind != tok.EOF {
                    if p.cur.Kind == tok.STRING {
                        path := unquote(p.cur.Lexeme)
                        f.Imports = append(f.Imports, path)
                        f.Decls = append(f.Decls, astpkg.ImportDecl{Path: path})
                        f.Stmts = append(f.Stmts, astpkg.ImportDecl{Path: path})
                        p.next()
                        continue
                    }
                    if p.cur.Kind == tok.RPAREN {
                        p.next()
                        break
                    }
                    // optional alias before string
                    if p.cur.Kind == tok.IDENT {
                        alias := p.cur.Lexeme
                        p.next()
                        if p.cur.Kind == tok.STRING {
                            path := unquote(p.cur.Lexeme)
                            f.Imports = append(f.Imports, path)
                            f.Decls = append(f.Decls, astpkg.ImportDecl{Path: path, Alias: alias})
                            f.Stmts = append(f.Stmts, astpkg.ImportDecl{Path: path, Alias: alias})
                            p.next()
                            continue
                        }
                    }
                    // skip
                    p.next()
                }
                continue
            }
            p.errorf("invalid import declaration")
            p.synchronize()
            continue
        }
        // func declaration scaffold: func IDENT (...) { ... }
        if p.cur.Kind == tok.KW_FUNC {
            p.next()
            name := ""
            if p.cur.Kind == tok.IDENT { name = p.cur.Lexeme; p.next() } else { p.errorf("expected function name") }
            // params
            if p.cur.Kind == tok.LPAREN {
                depth := 1; p.next()
                for depth > 0 && p.cur.Kind != tok.EOF {
                    if p.cur.Kind == tok.LPAREN { depth++ }
                    if p.cur.Kind == tok.RPAREN { depth--; if depth==0 { p.next(); break } }
                    p.next()
                }
            }
            // optional result list (skip naive)
            if p.cur.Kind == tok.LPAREN { // tuple
                depth := 1; p.next()
                for depth > 0 && p.cur.Kind != tok.EOF {
                    if p.cur.Kind == tok.LPAREN { depth++ }
                    if p.cur.Kind == tok.RPAREN { depth--; if depth==0 { p.next(); break } }
                    p.next()
                }
            }
            // body block
            if p.cur.Kind == tok.LBRACE {
                depth := 1; p.next()
                for depth > 0 && p.cur.Kind != tok.EOF {
                    if p.cur.Kind == tok.LBRACE { depth++ }
                    if p.cur.Kind == tok.RBRACE { depth--; if depth==0 { p.next(); break } }
                    p.next()
                }
            } else { p.errorf("expected function body") }
            fd := astpkg.FuncDecl{Name: name}
            f.Decls = append(f.Decls, fd)
            f.Stmts = append(f.Stmts, fd)
            if p.cur.Kind != tok.SEMI && p.cur.Kind != tok.KW_FUNC {
                // attempt to move forward to next statement
                p.synchronize()
            }
            continue
        }
        // Keep unparsed token as Bad node for now
        f.Stmts = append(f.Stmts, astpkg.Bad{Tok: p.cur})
        p.next()
    }
    return f
}

// parsePipelineDecl parses: pipeline IDENT { Node(args) ('.'|'->') Node(args) ... }
func (p *Parser) parsePipelineDecl() astpkg.PipelineDecl {
    // consume 'pipeline'
    p.next()
    name := ""
    if p.cur.Kind == tok.IDENT { name = p.cur.Lexeme; p.next() }
    // require '{'
    if p.cur.Kind != tok.LBRACE { return astpkg.PipelineDecl{} }
    p.next()
    steps, connectors := p.parseNodeChain()
    // require '}'
    if p.cur.Kind == tok.RBRACE { p.next() }
    return astpkg.PipelineDecl{Name: name, Steps: steps, Connectors: connectors}
}

// parseNodeChain parses Node(args) ('.'|'->') Node(args) ... until '}' or EOF
func (p *Parser) parseNodeChain() ([]astpkg.NodeCall, []string) {
    steps := []astpkg.NodeCall{}
    connectors := []string{}
    // Expect first node
    n, ok := p.parseNodeCall()
    if !ok { return steps, connectors }
    steps = append(steps, n)
    // Zero or more (. or ->) NodeCall
    for p.cur.Kind == tok.DOT || p.cur.Kind == tok.ARROW {
        conn := "."
        if p.cur.Kind == tok.ARROW { conn = "->" }
        p.next()
        n2, ok := p.parseNodeCall()
        if !ok { break }
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
    name := p.cur.Lexeme
    if name == "" { // for keyword tokens, Lexeme may carry source; fall back to kind name
        switch p.cur.Kind {
        case tok.KW_INGRESS: name = "ingress"
        case tok.KW_TRANSFORM: name = "transform"
        case tok.KW_FANOUT: name = "fanout"
        case tok.KW_COLLECT: name = "collect"
        case tok.KW_EGRESS: name = "egress"
        }
    }
    p.next()
    if p.cur.Kind != tok.LPAREN { return astpkg.NodeCall{Name: name}, true }
    // parse arguments as raw strings; handle nesting and commas at depth=1
    p.next() // consume '('
    args := []string{}
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
                if s != "" { args = append(args, s) }
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
                buf.Reset()
                p.next()
                continue
            }
            buf.WriteString(",")
            p.next()
        default:
            if p.cur.Lexeme != "" { buf.WriteString(p.cur.Lexeme) } else { buf.WriteRune(rune(p.cur.Kind)) }
            p.next()
        }
    }
    return astpkg.NodeCall{Name: name, Args: args}, true
}

// ExtractImports finds import paths in a minimal Go-like syntax:
//
//	import "path"
//	import (
//	  "a"
//	  "b"
//	)
func ExtractImports(src string) []string {
	p := New(src)
	f := p.ParseFile()
	out := make([]string, len(f.Imports))
	copy(out, f.Imports)
	return out
}

func unquote(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s[1 : len(s)-1]
	}
	return s
}
