package parser

import (
	"strings"
	"unicode"

	astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
	"github.com/sam-caldwell/ami/src/ami/compiler/diag"
	scan "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
	tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

type Parser struct {
	s      *scan.Scanner
	cur    tok.Token
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

func (p *Parser) posFrom(t tok.Token) astpkg.Position {
	return astpkg.Position{Line: t.Line, Column: t.Column, Offset: t.Offset}
}

func (p *Parser) consumeComments() []astpkg.Comment {
	cs := p.s.ConsumeComments()
	if len(cs) == 0 {
		return nil
	}
	out := make([]astpkg.Comment, 0, len(cs))
	for _, c := range cs {
		out = append(out, astpkg.Comment{Text: c.Text, Pos: astpkg.Position{Line: c.Line, Column: c.Column, Offset: c.Offset}})
	}
	return out
}

func (p *Parser) errorf(msg string) {
	p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PARSE", Message: msg, File: p.file})
}

// synchronize advances the parser until a likely statement boundary is found.
func (p *Parser) synchronize() {
	for p.cur.Kind != tok.EOF {
		if p.cur.Kind == tok.SEMI {
			p.next()
			return
		}
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
		// pragma directives
		if p.cur.Kind == tok.PRAGMA {
			pending := p.consumeComments()
			// parse: <name> <payload...>
			parts := strings.Fields(p.cur.Lexeme)
			name := ""
			payload := strings.TrimSpace(p.cur.Lexeme)
			if len(parts) > 0 {
				name = parts[0]
				payload = strings.TrimSpace(strings.TrimPrefix(p.cur.Lexeme, parts[0]))
			}
			d := astpkg.Directive{Name: name, Payload: strings.TrimSpace(payload), Pos: p.posFrom(p.cur), Comments: pending}
			f.Directives = append(f.Directives, d)
			p.next()
			continue
		}
		// package clause: package IDENT [ ':' version ]
		if p.cur.Kind == tok.KW_PACKAGE || (p.cur.Kind == tok.IDENT && p.cur.Lexeme == "package") {
			p.next()
			if p.cur.Kind == tok.IDENT {
				f.Package = p.cur.Lexeme
				// validate package ident per 6.1
				if !ValidatePackageIdent(f.Package) {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_PACKAGE", Message: "invalid package identifier", File: p.file})
				}
				// disallow blank identifier as package name
				if f.Package == "_" {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_PACKAGE_BLANK", Message: "blank identifier '_' cannot be used as package name", File: p.file})
				}
				p.next()
				// optional version after ':' consisting of IDENT/NUMBER/DOT/MINUS/PLUS tokens
				if p.cur.Kind == tok.COLON {
					// consume ':' and read contiguous non-space run from source as version
					p.next()
					src := p.s.Source()
					start := p.cur.Offset
					end := start
					for end < len(src) {
						c := src[end]
						if c == ' ' || c == tok.LexTab || c == tok.LexCr || c == tok.LexLf {
							break
						}
						end++
					}
					ver := strings.TrimSpace(src[start:end])
					if ver != "" {
						f.Version = strings.TrimSpace(ver)
						if !ValidateVersion(ver) {
							p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_PACKAGE_VERSION", Message: "invalid package version (expected SemVer)", File: p.file})
						}
					}
					// advance tokens to first token at or beyond 'end'
					for p.cur.Kind != tok.EOF && p.cur.Offset < end {
						p.next()
					}
				}
				continue
			}
			// if not IDENT, report bad package name
			p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_PACKAGE", Message: "invalid package identifier", File: p.file})
			p.synchronize()
		}
		// pipeline declaration: pipeline IDENT { <chain> }
		if p.cur.Kind == tok.KW_PIPELINE {
			pending := p.consumeComments()
			start := p.posFrom(p.cur)
			decl := p.parsePipelineDecl()
			decl.Pos = start
			decl.Comments = pending
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
			pending := p.consumeComments()
			start := p.posFrom(p.cur)
			p.next()
			// import "path"
			if p.cur.Kind == tok.STRING {
				path := unquote(p.cur.Lexeme)
				// validate import path
				if !ValidateImportPath(path) {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
				}
				f.Imports = append(f.Imports, path)
				id := astpkg.ImportDecl{Path: path, Pos: start, Comments: pending}
				f.Decls = append(f.Decls, id)
				f.Stmts = append(f.Stmts, id)
				p.next()
				continue
			}
			// import alias "path" -> skip alias
			if p.cur.Kind == tok.IDENT {
				alias := p.cur.Lexeme
				if alias == "_" {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IMPORT_ALIAS", Message: "blank identifier '_' cannot be used as import alias", File: p.file})
				}
				// Lookahead for quoted path form
				save := p.cur
				p.next()
				if p.cur.Kind == tok.STRING { // alias + quoted path
					path := unquote(p.cur.Lexeme)
					if alias == "_" {
						p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IMPORT_ALIAS", Message: "blank identifier '_' cannot be used as import alias", File: p.file})
					}
					if !ValidateImportPath(path) {
						p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
					}
					f.Imports = append(f.Imports, path)
					id := astpkg.ImportDecl{Path: path, Alias: alias, Pos: start, Comments: pending}
					f.Decls = append(f.Decls, id)
					f.Stmts = append(f.Stmts, id)
					p.next()
					continue
				}
				// Not quoted: treat alias token as first segment of unquoted path
				if path, cons, ok := p.parseImportUnquotedWithPrefix(save.Lexeme); ok {
					if !ValidateImportPath(path) {
						p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
					}
					if cons != "" && !ValidateImportConstraint(cons) {
						p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_CONSTRAINT", Message: "invalid import constraint", File: p.file})
					}
					f.Imports = append(f.Imports, path)
					id := astpkg.ImportDecl{Path: path, Constraint: cons, Pos: start, Comments: pending}
					f.Decls = append(f.Decls, id)
					f.Stmts = append(f.Stmts, id)
					continue
				}
			}
			// import ( ... )
			if p.cur.Kind == tok.LPAREN {
				p.next()
				for p.cur.Kind != tok.EOF {
					if p.cur.Kind == tok.STRING {
						path := unquote(p.cur.Lexeme)
						if !ValidateImportPath(path) {
							p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
						}
						f.Imports = append(f.Imports, path)
						id := astpkg.ImportDecl{Path: path, Pos: start, Comments: pending}
						f.Decls = append(f.Decls, id)
						f.Stmts = append(f.Stmts, id)
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
						if alias == "_" {
							p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IMPORT_ALIAS", Message: "blank identifier '_' cannot be used as import alias", File: p.file})
						}
						p.next()
						if p.cur.Kind == tok.STRING {
							path := unquote(p.cur.Lexeme)
							if !ValidateImportPath(path) {
								p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
							}
							f.Imports = append(f.Imports, path)
							id := astpkg.ImportDecl{Path: path, Alias: alias, Pos: start, Comments: pending}
							f.Decls = append(f.Decls, id)
							f.Stmts = append(f.Stmts, id)
							p.next()
							continue
						}
						// else: alias not followed by string; parse unquoted using alias as prefix
						if path, cons, ok := p.parseImportUnquotedWithPrefix(alias); ok {
							if !ValidateImportPath(path) {
								p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
							}
							if cons != "" && !ValidateImportConstraint(cons) {
								p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_CONSTRAINT", Message: "invalid import constraint", File: p.file})
							}
							f.Imports = append(f.Imports, path)
							id := astpkg.ImportDecl{Path: path, Constraint: cons, Pos: start, Comments: pending}
							f.Decls = append(f.Decls, id)
							f.Stmts = append(f.Stmts, id)
							continue
						}
					}
					// unquoted path (with optional constraint)
					if path, cons, ok := p.parseImportUnquoted(); ok {
						if !ValidateImportPath(path) {
							p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
						}
						if cons != "" && !ValidateImportConstraint(cons) {
							p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_CONSTRAINT", Message: "invalid import constraint", File: p.file})
						}
						f.Imports = append(f.Imports, path)
						id := astpkg.ImportDecl{Path: path, Constraint: cons, Pos: start, Comments: pending}
						f.Decls = append(f.Decls, id)
						f.Stmts = append(f.Stmts, id)
						continue
					}
					// skip
					p.next()
				}
				continue
			}
			// unquoted single import (with optional constraint)
			if path, cons, ok := p.parseImportUnquoted(); ok {
				if !ValidateImportPath(path) {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BAD_IMPORT", Message: "invalid import path", File: p.file})
				}
				if cons != "" && !ValidateImportConstraint(cons) {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_IMPORT_CONSTRAINT", Message: "invalid import constraint", File: p.file})
				}
				f.Imports = append(f.Imports, path)
				id := astpkg.ImportDecl{Path: path, Constraint: cons, Pos: start, Comments: pending}
				f.Decls = append(f.Decls, id)
				f.Stmts = append(f.Stmts, id)
				continue
			}
			p.errorf("invalid import declaration")
			p.synchronize()
			continue
		}
		// enum declaration: enum IDENT { IDENT [= value] (, IDENT [= value]) ... }
		if p.cur.Kind == tok.KW_ENUM {
			pending := p.consumeComments()
			start := p.posFrom(p.cur)
			ed := p.parseEnumDecl()
			ed.Pos = start
			ed.Comments = pending
			if ed.Name != "" {
				f.Decls = append(f.Decls, ed)
				f.Stmts = append(f.Stmts, ed)
				continue
			}
			p.errorf("invalid enum declaration")
			p.synchronize()
			continue
		}
		// struct declaration: struct IDENT { IDENT Type [,|;] ... }
		if p.cur.Kind == tok.KW_STRUCT {
			pending := p.consumeComments()
			start := p.posFrom(p.cur)
			sd := p.parseStructDecl()
			sd.Pos = start
			sd.Comments = pending
			if sd.Name != "" {
				f.Decls = append(f.Decls, sd)
				f.Stmts = append(f.Stmts, sd)
				continue
			}
			p.errorf("invalid struct declaration")
			p.synchronize()
			continue
		}
		// func declaration: func IDENT (params) [result] { ... }
		if p.cur.Kind == tok.KW_FUNC {
			pending := p.consumeComments()
			start := p.posFrom(p.cur)
			p.next()
			name := ""
			if p.cur.Kind == tok.IDENT {
				name = p.cur.Lexeme
				if name == "_" {
					p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_BLANK_IDENT_ILLEGAL", Message: "blank identifier '_' cannot be used as a function name", File: p.file})
				}
				p.next()
			} else {
				p.errorf("expected function name")
			}
			// params
			var params []astpkg.Param
			var results []astpkg.TypeRef
			if p.cur.Kind == tok.LPAREN {
				p.next()
				params = p.parseParamList()
				if p.cur.Kind == tok.RPAREN {
					p.next()
				}
			}
			// optional result list or single type
			if p.cur.Kind == tok.LPAREN { // tuple
				p.next()
				results = p.parseResultList()
				if p.cur.Kind == tok.RPAREN {
					p.next()
				}
			} else {
				if tr, ok := p.parseType(); ok {
					results = append(results, tr)
				}
			}
			// body block: capture tokens and build a simple statement AST (scaffold)
			var body []tok.Token
			var bodyStmts []astpkg.Stmt
			bodyComments := make(map[int][]astpkg.Comment)
			if p.cur.Kind == tok.LBRACE {
				depth := 1
				p.next()
				// collect tokens inside body
				for depth > 0 && p.cur.Kind != tok.EOF {
					// collect any comments preceding this token and associate to token offset
					if pcs := p.consumeComments(); len(pcs) > 0 {
						bodyComments[p.cur.Offset] = append(bodyComments[p.cur.Offset], pcs...)
					}
					body = append(body, p.cur)
					// address-of is not allowed (2.3.2)
					if p.cur.Kind == tok.AMP {
						p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'&' address-of operator is not allowed; AMI does not expose raw pointers (see 2.3.2)", File: p.file})
					}
					if p.cur.Kind == tok.LBRACE {
						depth++
					}
					if p.cur.Kind == tok.RBRACE {
						depth--
						if depth == 0 {
							p.next()
							break
						}
					}
					p.next()
				}
				// build simple statement list from captured tokens
				bodyStmts = parseBodyStmts(body, bodyComments)
			} else {
				p.errorf("expected function body")
			}
			fd := astpkg.FuncDecl{Name: name, Params: params, Result: results, Body: body, BodyStmts: bodyStmts, Pos: start, Comments: pending}
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

// parseEnumDecl parses: enum IDENT '{' members '}'
// member: IDENT [ '=' (NUMBER | STRING) ]
func (p *Parser) parseEnumDecl() astpkg.EnumDecl {
	// consume 'enum'
	p.next()
	name := ""
	if p.cur.Kind == tok.IDENT {
		name = p.cur.Lexeme
		p.next()
	} else {
		return astpkg.EnumDecl{}
	}
	if p.cur.Kind != tok.LBRACE {
		return astpkg.EnumDecl{}
	}
	p.next() // consume '{'
	var members []astpkg.EnumMember
	for p.cur.Kind != tok.EOF {
		if p.cur.Kind == tok.RBRACE {
			p.next()
			break
		}
		// skip stray commas
		if p.cur.Kind == tok.COMMA {
			p.next()
			continue
		}
		if p.cur.Kind != tok.IDENT {
			// skip until next comma or '}'
			p.next()
			continue
		}
		memName := p.cur.Lexeme
		p.next()
		memVal := ""
		if p.cur.Kind == tok.ASSIGN {
			p.next()
			switch p.cur.Kind {
			case tok.STRING:
				// preserve quotes so semantics can distinguish string vs number
				memVal = p.cur.Lexeme
				p.next()
			case tok.NUMBER:
				memVal = p.cur.Lexeme
				p.next()
			case tok.MINUS:
				// allow negative numbers
				p.next()
				if p.cur.Kind == tok.NUMBER {
					memVal = "-" + p.cur.Lexeme
					p.next()
				}
			default:
				// unknown value; leave empty to be caught by semantics if needed
			}
		}
		members = append(members, astpkg.EnumMember{Name: memName, Value: memVal})
		if p.cur.Kind == tok.COMMA {
			p.next()
		}
	}
	return astpkg.EnumDecl{Name: name, Members: members}
}

// parseStructDecl parses: struct IDENT '{' fields '}'
// field := IDENT TypeRef
// separators: optional comma or semicolon between fields
func (p *Parser) parseStructDecl() astpkg.StructDecl {
	p.next() // consume 'struct'
	name := ""
	if p.cur.Kind == tok.IDENT {
		name = p.cur.Lexeme
		p.next()
	} else {
		return astpkg.StructDecl{}
	}
	if p.cur.Kind != tok.LBRACE {
		return astpkg.StructDecl{}
	}
	p.next() // consume '{'
	var fields []astpkg.Field
	for p.cur.Kind != tok.EOF {
		if p.cur.Kind == tok.RBRACE {
			p.next()
			break
		}
		if p.cur.Kind == tok.COMMA || p.cur.Kind == tok.SEMI {
			p.next()
			continue
		}
		if p.cur.Kind != tok.IDENT {
			// skip token and continue
			p.next()
			continue
		}
		fname := p.cur.Lexeme
		p.next()
		ftype, ok := p.parseType()
		if !ok {
			fields = append(fields, astpkg.Field{Name: fname, Type: astpkg.TypeRef{}})
		} else {
			fields = append(fields, astpkg.Field{Name: fname, Type: ftype})
		}
		if p.cur.Kind == tok.COMMA || p.cur.Kind == tok.SEMI {
			p.next()
		}
	}
	return astpkg.StructDecl{Name: name, Fields: fields}
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

// ImportItem describes an import with optional alias/constraint from source.
type ImportItem struct {
	Path       string
	Alias      string
	Constraint string
}

// ExtractImportItems returns detailed import items with optional constraints.
func ExtractImportItems(src string) []ImportItem {
	p := New(src)
	f := p.ParseFile()
	var out []ImportItem
	for _, d := range f.Decls {
		if id, ok := d.(astpkg.ImportDecl); ok {
			out = append(out, ImportItem{Path: id.Path, Alias: id.Alias, Constraint: id.Constraint})
		}
	}
	return out
}

func unquote(s string) string {
	if len(s) >= 2 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s[1 : len(s)-1]
	}
	return s
}

// --- Simple body statement parsing from captured tokens ---

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

// parseImportUnquoted parses an unquoted import path optionally followed by a constraint.
// It accepts module paths like github.com/org/repo and ami/stdlib/io and a constraint
// in the form ">= vX.Y.Z" (operator and version may be separated by spaces).
func (p *Parser) parseImportUnquoted() (string, string, bool) {
	return p.parseImportUnquotedWithPrefix("")
}

func (p *Parser) parseImportUnquotedWithPrefix(prefix string) (string, string, bool) {
	var b strings.Builder
	if prefix != "" {
		b.WriteString(prefix)
	} else {
		if !(p.cur.Kind == tok.IDENT) {
			return "", "", false
		}
		b.WriteString(p.cur.Lexeme)
		p.next()
	}
	// accumulate path segments: ('.' IDENT | '/' IDENT | '-' in segment) etc.
	for {
		switch p.cur.Kind {
		case tok.DOT:
			b.WriteByte('.')
			p.next()
			if p.cur.Kind == tok.IDENT {
				b.WriteString(p.cur.Lexeme)
				p.next()
			}
			continue
		case tok.SLASH:
			b.WriteByte('/')
			p.next()
			if p.cur.Kind == tok.IDENT {
				b.WriteString(p.cur.Lexeme)
				p.next()
			}
			continue
		case tok.MINUS:
			// minus inside a segment
			b.WriteByte('-')
			p.next()
			// allow continuing ident characters
			if p.cur.Kind == tok.IDENT {
				b.WriteString(p.cur.Lexeme)
				p.next()
			}
			continue
		default:
			// end of path
		}
		break
	}
	path := b.String()
	// optional constraint: read rest of line and accept forms starting with ">="
	cons := ""
	{
		src := p.s.Source()
		start := p.cur.Offset
		end := start
		for end < len(src) {
			if src[end] == '\n' {
				break
			}
			end++
		}
		tail := strings.TrimSpace(src[start:end])
		if strings.HasPrefix(tail, ">=") {
			cons = ">=" + " " + strings.TrimSpace(strings.TrimPrefix(tail, ">="))
			// advance tokens to the end of line to keep parser in sync
			for p.cur.Kind != tok.EOF && p.cur.Offset < end {
				p.next()
			}
		}
	}
	if path == "" {
		return "", "", false
	}
	return path, cons, true
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
	// helper to convert token position
	toPos := func(t tok.Token) astpkg.Position {
		return astpkg.Position{Line: t.Line, Column: t.Column, Offset: t.Offset}
	}
	// 'var' declaration: var IDENT [Type] [= expr]
	if bp.cur().Kind == tok.KW_VAR {
		start := bp.cur()
		bp.next()
		if bp.cur().Kind != tok.IDENT {
			return nil, false
		}
		nameTok := bp.cur()
		name := nameTok.Lexeme
		bp.next()
		// try parse optional type
		tr, hasType := bp.parseTypeRef()
		// optional initializer
		var init astpkg.Expr
		if bp.cur().Kind == tok.ASSIGN {
			bp.next()
			if x, ok := bp.parseExpr(); ok {
				init = x
			}
		}
		if hasType {
			return bp.attachStmtComments(start, astpkg.VarDeclStmt{Name: name, Type: tr, Init: init, Pos: toPos(start)}), true
		}
		return bp.attachStmtComments(start, astpkg.VarDeclStmt{Name: name, Init: init, Pos: toPos(start)}), true
	}
	// 'defer' statement: defer <expr>
	if bp.cur().Kind == tok.KW_DEFER {
		start := bp.cur()
		bp.next()
		if x, ok := bp.parseExpr(); ok {
			return bp.attachStmtComments(start, astpkg.DeferStmt{X: x, Pos: toPos(start)}), true
		}
		return bp.attachStmtComments(start, astpkg.DeferStmt{X: nil, Pos: toPos(start)}), true
	}
	// 'return' statement: return [expr]
	if bp.cur().Kind == tok.KW_RETURN {
		start := bp.cur()
		bp.next()
		// parse zero or more expressions separated by commas
		var results []astpkg.Expr
		if x, ok := bp.parseExpr(); ok {
			results = append(results, x)
			for bp.cur().Kind == tok.COMMA {
				bp.next()
				if y, ok2 := bp.parseExpr(); ok2 {
					results = append(results, y)
				} else {
					break
				}
			}
		}
		return bp.attachStmtComments(start, astpkg.ReturnStmt{Results: results, Pos: toPos(start)}), true
	}
	// Note: AMI does not support Rust-like `mut { ... }` blocks.
	// Any appearance of `mut` is treated as an identifier token in expressions.
	// assignment or call expr
	// Try parse LHS expr
	save := bp.i
	if lhs, ok := bp.parseExpr(); ok {
		if bp.cur().Kind == tok.ASSIGN {
			assignTok := bp.cur()
			bp.next()
			if rhs, ok2 := bp.parseExpr(); ok2 {
				return bp.attachStmtComments(bp.toks[save], astpkg.AssignStmt{LHS: lhs, RHS: rhs, Pos: toPos(assignTok)}), true
			}
			// rollback
			bp.i = save
		} else {
			// treat any standalone expression as a statement for semantic scans
			// position from the first token of the expression
			if start := save; start < len(bp.toks) {
				return bp.attachStmtComments(bp.toks[start], astpkg.ExprStmt{X: lhs, Pos: toPos(bp.toks[start])}), true
			}
			return bp.attachStmtComments(tok.Token{}, astpkg.ExprStmt{X: lhs}), true
		}
	}
	bp.i = save
	return nil, false
}

func (bp *bodyParser) parseExpr() (astpkg.Expr, bool) { return bp.parseBinaryExpr() }

func (bp *bodyParser) parseArgs() []astpkg.Expr {
	// assume current is '('
	if bp.cur().Kind != tok.LPAREN {
		return nil
	}
	bp.next()
	var out []astpkg.Expr
	depth := 1
	for !bp.atEnd() && depth > 0 {
		switch bp.cur().Kind {
		case tok.LPAREN:
			depth++
			bp.next()
		case tok.RPAREN:
			depth--
			bp.next()
			if depth == 0 {
				break
			}
		case tok.COMMA:
			bp.next()
		default:
			if e, ok := bp.parseExpr(); ok {
				out = append(out, e)
			} else {
				bp.next()
			}
		}
	}
	return out
}

func (bp *bodyParser) attachStmtComments(start tok.Token, s astpkg.Stmt) astpkg.Stmt {
	var off int
	if start.Kind != tok.EOF {
		off = start.Offset
	} else if bp.i < len(bp.toks) {
		off = bp.toks[bp.i].Offset
	}
	if off != 0 {
		if cmts, ok := bp.comments[off]; ok {
			switch v := s.(type) {
			case astpkg.AssignStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			case astpkg.ExprStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			case astpkg.DeferStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			case astpkg.ReturnStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			case astpkg.VarDeclStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			case astpkg.BlockStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			case astpkg.MutBlockStmt:
				v.Comments = append(v.Comments, cmts...)
				return v
			}
		}
	}
	return s
}

// --- Binary expression parsing (precedence-climbing) ---

func (bp *bodyParser) parseBinaryExpr() (astpkg.Expr, bool) {
	left, ok := bp.parsePrimary()
	if !ok {
		return nil, false
	}
	return bp.parseBinaryRHS(0, left)
}

func (bp *bodyParser) parsePrimary() (astpkg.Expr, bool) {
	t := bp.cur()
	toPos := func(tk tok.Token) astpkg.Position {
		return astpkg.Position{Line: tk.Line, Column: tk.Column, Offset: tk.Offset}
	}
	switch t.Kind {
	case tok.KW_SLICE, tok.KW_SET:
		// slice<T>{...} or set<T>{...}
		kind := "slice"
		if t.Kind == tok.KW_SET {
			kind = "set"
		}
		start := t
		bp.next()
		var args []astpkg.TypeRef
		if bp.cur().Kind == tok.LT {
			bp.next()
			if a, ok := bp.parseTypeRef(); ok {
				args = append(args, a)
			}
			if bp.cur().Kind == tok.GT {
				bp.next()
			}
		}
		// expect '{'
		if bp.cur().Kind != tok.LBRACE {
			return nil, false
		}
		bp.next()
		var elems []astpkg.Expr
		for !bp.atEnd() && bp.cur().Kind != tok.RBRACE {
			if bp.cur().Kind == tok.COMMA {
				bp.next()
				continue
			}
			if e, ok := bp.parseExpr(); ok {
				elems = append(elems, e)
			} else {
				bp.next()
			}
			if bp.cur().Kind == tok.COMMA {
				bp.next()
			}
		}
		if bp.cur().Kind == tok.RBRACE {
			bp.next()
		}
		return astpkg.ContainerLit{Kind: kind, TypeArgs: args, Elems: elems, Pos: toPos(start)}, true
	case tok.KW_MAP:
		// map<K,V>{k:v,...}
		start := t
		bp.next()
		var args []astpkg.TypeRef
		if bp.cur().Kind == tok.LT {
			bp.next()
			if a, ok := bp.parseTypeRef(); ok {
				args = append(args, a)
			}
			if bp.cur().Kind == tok.COMMA {
				bp.next()
			}
			if b, ok := bp.parseTypeRef(); ok {
				args = append(args, b)
			}
			if bp.cur().Kind == tok.GT {
				bp.next()
			}
		}
		if bp.cur().Kind != tok.LBRACE {
			return nil, false
		}
		bp.next()
		var kvs []astpkg.MapElem
		for !bp.atEnd() && bp.cur().Kind != tok.RBRACE {
			if bp.cur().Kind == tok.COMMA {
				bp.next()
				continue
			}
			k, ok := bp.parseExpr()
			if !ok {
				bp.next()
				continue
			}
			if bp.cur().Kind != tok.COLON {
				bp.next()
				continue
			}
			bp.next()
			v, ok2 := bp.parseExpr()
			if !ok2 {
				bp.next()
				continue
			}
			kvs = append(kvs, astpkg.MapElem{Key: k, Value: v})
			if bp.cur().Kind == tok.COMMA {
				bp.next()
			}
		}
		if bp.cur().Kind == tok.RBRACE {
			bp.next()
		}
		return astpkg.ContainerLit{Kind: "map", TypeArgs: args, MapElems: kvs, Pos: toPos(start)}, true
	case tok.IDENT:
		// method call: recv . method (args)
		// or simple call: name(args)
		name := t.Lexeme
		bp.next()
		// method selector
		if bp.cur().Kind == tok.DOT {
			recv := astpkg.Ident{Name: name, Pos: toPos(t)}
			bp.next()
			if bp.cur().Kind == tok.IDENT {
				sel := bp.cur().Lexeme
				selTok := bp.cur()
				bp.next()
				if bp.cur().Kind == tok.LPAREN {
					args := bp.parseArgs()
					return astpkg.CallExpr{Fun: astpkg.SelectorExpr{X: recv, Sel: sel, Pos: toPos(selTok)}, Args: args, Pos: toPos(t)}, true
				}
				return astpkg.SelectorExpr{X: recv, Sel: sel, Pos: toPos(selTok)}, true
			}
			return astpkg.Ident{Name: name, Pos: toPos(t)}, true
		}
		if bp.cur().Kind == tok.LPAREN {
			args := bp.parseArgs()
			return astpkg.CallExpr{Fun: astpkg.Ident{Name: name, Pos: toPos(t)}, Args: args, Pos: toPos(t)}, true
		}
		return astpkg.Ident{Name: name, Pos: toPos(t)}, true
	case tok.STRING:
		bp.next()
		return astpkg.BasicLit{Kind: "string", Value: t.Lexeme, Pos: toPos(t)}, true
	case tok.NUMBER:
		bp.next()
		return astpkg.BasicLit{Kind: "number", Value: t.Lexeme, Pos: toPos(t)}, true
	case tok.STAR:
		bp.next()
		if x, ok := bp.parsePrimary(); ok {
			return astpkg.UnaryExpr{Op: "*", X: x, Pos: toPos(t)}, true
		}
		return nil, false
	case tok.AMP:
		bp.next()
		if x, ok := bp.parsePrimary(); ok {
			return astpkg.UnaryExpr{Op: "&", X: x, Pos: toPos(t)}, true
		}
		return nil, false
	case tok.LPAREN:
		bp.next()
		e, ok := bp.parseBinaryExpr()
		if bp.cur().Kind == tok.RPAREN {
			bp.next()
		}
		return e, ok
	default:
		return nil, false
	}
}

func (bp *bodyParser) precedence(k tok.Kind) int {
	switch k {
	case tok.STAR, tok.SLASH, tok.PERCENT:
		return 40
	case tok.PLUS, tok.MINUS:
		return 30
	case tok.LT, tok.LTE, tok.GT, tok.GTE:
		return 20
	case tok.EQ, tok.NEQ:
		return 10
	default:
		return -1
	}
}

func (bp *bodyParser) parseBinaryRHS(minPrec int, left astpkg.Expr) (astpkg.Expr, bool) {
    for {
        // defensive: ensure forward progress to avoid pathological loops
        beforeI := bp.i
        opTok := bp.cur()
        prec := bp.precedence(opTok.Kind)
        if prec < minPrec {
            break
        }
		// treat '*' as multiplication only when not the LHS mutation marker: pattern '*' IDENT '='
		if opTok.Kind == tok.STAR {
			if bp.i+2 < len(bp.toks) && bp.toks[bp.i+1].Kind == tok.IDENT && bp.toks[bp.i+2].Kind == tok.ASSIGN {
				break
			}
		}
		// consume operator
		bp.next()
		// parse right operand
		right, ok := bp.parsePrimary()
		if !ok {
			return left, true
		}
        // If next operator has higher precedence, parse it first
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
                if !ok2 {
                    break
                }
            } else {
                break
            }
        }
		op := opTok.Lexeme
		if op == "" {
			switch opTok.Kind {
			case tok.STAR:
				op = "*"
			case tok.SLASH:
				op = "/"
			case tok.PERCENT:
				op = "%"
			case tok.PLUS:
				op = "+"
			case tok.MINUS:
				op = "-"
			case tok.LT:
				op = "<"
			case tok.LTE:
				op = "<="
			case tok.GT:
				op = ">"
			case tok.GTE:
				op = ">="
			case tok.EQ:
				op = "=="
			case tok.NEQ:
				op = "!="
			}
		}
        left = astpkg.BinaryExpr{X: left, Op: op, Y: right}
        if bp.i == beforeI { // no progress; bail out to avoid infinite loop
            break
        }
    }
    return left, true
}

// parseTypeRef parses a simple TypeRef in function bodies, mirroring Parser.parseType
// sufficiently for local variable declarations. Rejects pointer type syntax per AMI 2.3.2.
func (bp *bodyParser) parseTypeRef() (astpkg.TypeRef, bool) {
	var tr astpkg.TypeRef
	// reject pointer '*' proactively in type position
	if bp.cur().Kind == tok.STAR {
		bp.next()
		return tr, false
	}
	// slice []
	if bp.cur().Kind == tok.LBRACK {
		bp.next()
		if bp.cur().Kind == tok.RBRACK {
			tr.Slice = true
			bp.next()
		} else {
			return tr, false
		}
	}
	// type name
	switch bp.cur().Kind {
	case tok.IDENT, tok.KW_MAP, tok.KW_SET, tok.KW_SLICE:
		tr.Name = bp.cur().Lexeme
		bp.next()
	default:
		return tr, false
	}
	// generics <...>
	if bp.cur().Kind == tok.LT {
		bp.next()
		for !bp.atEnd() {
			if bp.cur().Kind == tok.GT {
				bp.next()
				break
			}
			if bp.cur().Kind == tok.COMMA {
				bp.next()
				continue
			}
			if arg, ok := bp.parseTypeRef(); ok {
				tr.Args = append(tr.Args, arg)
				continue
			}
			bp.next()
		}
	}
	return tr, true
}

// parseWorkerRef tries to interpret an argument string as a worker/factory reference.
func parseWorkerRef(s string) (astpkg.WorkerRef, bool) {
	// name or name(...)
	name := s
	hasCall := false
	if i := strings.IndexRune(s, '('); i >= 0 {
		name = strings.TrimSpace(s[:i])
		hasCall = true
	}
	if !isIdentLexeme(name) {
		return astpkg.WorkerRef{}, false
	}
	kind := "function"
	if hasCall || strings.HasPrefix(name, "New") {
		kind = "factory"
	}
	return astpkg.WorkerRef{Name: name, Kind: kind}, true
}

func isIdentLexeme(s string) bool {
	if s == "" {
		return false
	}
	r0 := []rune(s)[0]
	if !(unicode.IsLetter(r0) || r0 == '_') {
		return false
	}
	for _, r := range s[1:] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

// parseParamList parses zero or more parameters until ')'
func (p *Parser) parseParamList() []astpkg.Param {
	var params []astpkg.Param
	for p.cur.Kind != tok.RPAREN && p.cur.Kind != tok.EOF {
		if p.cur.Kind == tok.COMMA {
			p.next()
			continue
		}
		var name string
		if p.cur.Kind == tok.IDENT {
			ident := p.cur.Lexeme
			// try to parse type after ident
			// save position by making a copy of parser state is complex; instead try parseType and if fails treat as type-only
			p.next()
			// AMI 2.3.2: reject pointer type syntax proactively
			if p.cur.Kind == tok.STAR {
				p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'*' pointer type/dereference is not allowed; AMI does not expose raw pointers (see 2.3.2)", File: p.file})
				p.next()
			}
			if tr, ok := p.parseType(); ok {
				name = ident
				params = append(params, astpkg.Param{Name: name, Type: tr})
			} else {
				// ident was a type name
				params = append(params, astpkg.Param{Name: "", Type: astpkg.TypeRef{Name: ident}})
			}
		} else {
			if tr, ok := p.parseType(); ok {
				params = append(params, astpkg.Param{Name: "", Type: tr})
			} else {
				p.next()
			}
		}
		if p.cur.Kind == tok.COMMA {
			p.next()
		}
	}
	return params
}

func (p *Parser) parseResultList() []astpkg.TypeRef {
	var results []astpkg.TypeRef
	for p.cur.Kind != tok.RPAREN && p.cur.Kind != tok.EOF {
		if p.cur.Kind == tok.COMMA {
			p.next()
			continue
		}
		// reject pointer type syntax proactively
		if p.cur.Kind == tok.STAR {
			p.errors = append(p.errors, diag.Diagnostic{Level: diag.Error, Code: "E_PTR_UNSUPPORTED_SYNTAX", Message: "'*' pointer type/dereference is not allowed; AMI does not expose raw pointers (see 2.3.2)", File: p.file})
			p.next()
		}
		if tr, ok := p.parseType(); ok {
			results = append(results, tr)
		} else {
			p.next()
		}
		if p.cur.Kind == tok.COMMA {
			p.next()
		}
	}
	return results
}

// parseType parses '*'? '[]'? IDENT|KW_MAP|KW_SET|KW_SLICE ('<' Type {',' Type } '>')?
func (p *Parser) parseType() (astpkg.TypeRef, bool) {
	var tr astpkg.TypeRef
	// record starting offset
	start := p.cur.Offset
	// pointer
	if p.cur.Kind == tok.STAR {
		tr.Ptr = true
		p.next()
		start = p.cur.Offset
	}
	// slice []
	if p.cur.Kind == tok.LBRACK {
		start = p.cur.Offset
		p.next()
		if p.cur.Kind == tok.RBRACK {
			tr.Slice = true
			p.next()
		}
	}
	// accept identifiers and certain keywords as type names
	switch p.cur.Kind {
	case tok.IDENT, tok.KW_MAP, tok.KW_SET, tok.KW_SLICE:
		// ok
	default:
		return tr, false
	}
	tr.Name = p.cur.Lexeme
	tr.Offset = start
	p.next()
	// generics
	if p.cur.Kind == tok.LT {
		p.next()
		for p.cur.Kind != tok.EOF {
			if p.cur.Kind == tok.GT {
				p.next()
				break
			}
			if p.cur.Kind == tok.COMMA {
				p.next()
				continue
			}
			if arg, ok := p.parseType(); ok {
				tr.Args = append(tr.Args, arg)
				continue
			}
			p.next()
		}
	}
	return tr, true
}

// parseImportUnquotedWithPrefix is a tolerant placeholder to accept future
// unquoted import forms with optional version constraints. For now, disabled.
// (duplicate placeholders removed; real implementations exist above)
