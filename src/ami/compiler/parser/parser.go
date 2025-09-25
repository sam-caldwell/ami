package parser

import (
	"strings"

	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	scan "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

type Parser struct {
    s   *scan.Scanner
    cur tok.Token
}

func New(src string) *Parser {
	p := &Parser{s: scan.New(src)}
	p.next()
	return p
}

func (p *Parser) next() { p.cur = p.s.Next() }

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
        }
        // func declaration scaffold: func IDENT (...) { ... }
        if p.cur.Kind == tok.KW_FUNC {
            p.next()
            name := ""
            if p.cur.Kind == tok.IDENT { name = p.cur.Lexeme; p.next() }
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
            }
            fd := astpkg.FuncDecl{Name: name}
            f.Decls = append(f.Decls, fd)
            f.Stmts = append(f.Stmts, fd)
            continue
        }
        // Keep unparsed token as Bad node for now
        f.Stmts = append(f.Stmts, astpkg.Bad{Tok: p.cur})
        p.next()
    }
    return f
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
