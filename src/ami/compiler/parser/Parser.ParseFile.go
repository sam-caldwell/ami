package parser

import (
    "fmt"

    "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    "github.com/sam-caldwell/ami/src/ami/compiler/source"
    "github.com/sam-caldwell/ami/src/ami/compiler/token"
)

// ParseFile parses a single file, recognizing only a package declaration
// followed by imports and top-level declarations (func, pipeline, enum, error).
func (p *Parser) ParseFile() (*ast.File, error) {
    if p == nil {
        return nil, fmt.Errorf("nil parser")
    }
    f := &ast.File{}
    // Expect 'package' keyword
    if p.cur.Kind != token.KwPackage {
        p.errf("expected 'package', got %q", p.cur.Lexeme)
        p.syncTop()
        if p.cur.Kind != token.KwPackage {
            return f, p.firstErr()
        }
    }
    p.next()
    // Accept a standard identifier for the package name. Also allow certain
    // reserved keywords which are valid package identifiers in stdlib, e.g. "gpu".
    if p.cur.Kind != token.Ident && p.cur.Kind != token.KwGpu {
        p.errf("expected package name, got %q", p.cur.Lexeme)
        p.syncTop()
        if p.cur.Kind != token.Ident && p.cur.Kind != token.KwGpu {
            return f, p.firstErr()
        }
    }
    f.PackageName = p.cur.Lexeme
    f.PackagePos = p.cur.Pos
    p.next()

    // zero or more imports: import ident | alias "path" | import ( ... )
    for p.cur.Kind == token.KwImport {
        startPos := p.cur.Pos
        p.next()
        if p.cur.Kind == token.LParenSym {
            // block form: import ( line... )
            p.next()
            for p.cur.Kind != token.RParenSym && p.cur.Kind != token.EOF {
                if p.cur.Kind == token.SemiSym || p.cur.Kind == token.CommaSym {
                    p.next()
                    continue
                }
                if p.cur.Kind != token.Ident && p.cur.Kind != token.String {
                    p.errf("expected import path in block, got %q", p.cur.Lexeme)
                    p.next()
                    continue
                }
                alias := ""
                aliasPos := p.cur.Pos
                var path string
                var ppos source.Position
                if p.cur.Kind == token.String {
                    path = p.cur.Lexeme
                    ppos = p.cur.Pos
                    if len(path) >= 2 {
                        path = path[1 : len(path)-1]
                    }
                    p.next()
                } else if p.cur.Kind == token.Ident {
                    // Possible alias form: alias "path"
                    alias = p.cur.Lexeme
                    aliasPos = p.cur.Pos
                    p.next()
                    if p.cur.Kind == token.String {
                        path = p.cur.Lexeme
                        ppos = p.cur.Pos
                        if len(path) >= 2 {
                            path = path[1 : len(path)-1]
                        }
                        p.next()
                    } else {
                        // Treat previous ident as path (no alias)
                        path = alias
                        ppos = aliasPos
                        alias = ""
                    }
                }
                constraint := p.parseImportConstraint()
                im := &ast.ImportDecl{Pos: startPos, Path: path, Leading: p.pending, PathPos: ppos, Alias: alias, AliasPos: aliasPos, Constraint: constraint}
                p.pending = nil
                f.Decls = append(f.Decls, im)
            }
            if p.cur.Kind == token.RParenSym {
                p.next()
            } else {
                p.errf("missing ')' to close import block")
            }
            continue
        }
        // single-line form
        if p.cur.Kind != token.Ident && p.cur.Kind != token.String {
            p.errf("expected import path, got %q", p.cur.Lexeme)
            p.syncTop()
            continue
        }
        alias := ""
        aliasPos := p.cur.Pos
        var path string
        var ppos source.Position
        if p.cur.Kind == token.String {
            path = p.cur.Lexeme
            ppos = p.cur.Pos
            if len(path) >= 2 {
                path = path[1 : len(path)-1]
            }
            p.next()
        } else if p.cur.Kind == token.Ident {
            // Possible alias form
            alias = p.cur.Lexeme
            aliasPos = p.cur.Pos
            p.next()
            if p.cur.Kind == token.String {
                path = p.cur.Lexeme
                ppos = p.cur.Pos
                if len(path) >= 2 {
                    path = path[1 : len(path)-1]
                }
                p.next()
            } else {
                path = alias
                ppos = aliasPos
                alias = ""
            }
        }
        constraint := p.parseImportConstraint()
        im := &ast.ImportDecl{Pos: startPos, Path: path, Leading: p.pending, PathPos: ppos, Alias: alias, AliasPos: aliasPos, Constraint: constraint}
        p.pending = nil
        f.Decls = append(f.Decls, im)
    }

    // Top-level declarations in any order: decorators+func, pipeline, enum, error.
    for p.cur.Kind != token.EOF {
        switch p.cur.Kind {
        case token.AtSym, token.KwFunc:
            // collect any decorators in source order
            for p.cur.Kind == token.AtSym {
                if d, ok := p.parseDecorator(); ok {
                    p.pendingDecos = append(p.pendingDecos, d)
                } else {
                    // skip invalid decorator token
                    p.next()
                }
            }
            if p.cur.Kind != token.KwFunc {
                if len(p.pendingDecos) > 0 {
                    p.errf("decorators are only allowed immediately before function declarations")
                    p.pendingDecos = nil
                }
                // Not a function; re-evaluate outer loop for other decl kinds
                continue
            }
            fn, err := p.parseFuncDecl()
            if err != nil {
                p.errf("%v", err)
                p.syncTop()
            } else {
                f.Decls = append(f.Decls, fn)
            }
        case token.KwPipeline:
            pd, err := p.parsePipelineDecl()
            if err != nil {
                p.errf("%v", err)
                p.syncTop()
            } else {
                f.Decls = append(f.Decls, pd)
            }
        case token.KwEnum:
            ed, err := p.parseEnumDecl()
            if err != nil {
                p.errf("%v", err)
                p.syncTop()
            } else {
                f.Decls = append(f.Decls, ed)
            }
        case token.KwError:
            eb, err := p.parseErrorBlock()
            if err != nil {
                p.errf("%v", err)
                p.syncTop()
            } else {
                f.Decls = append(f.Decls, eb)
            }
        default:
            // Skip unknown/irrelevant tokens at top level
            p.next()
            continue
        }
    }
    // collect pragmas from raw file content (lines starting with '#pragma ')
    f.Pragmas = p.collectPragmas()
    return f, p.firstErr()
}
