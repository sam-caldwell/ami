package parser

import (
    "strings"

    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/diag"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// ParseFile parses the source into a lightweight AST file node.
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

