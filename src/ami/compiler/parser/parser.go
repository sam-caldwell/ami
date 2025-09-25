package parser

import (
    astpkg "github.com/sam-caldwell/ami/src/ami/compiler/ast"
    scan "github.com/sam-caldwell/ami/src/ami/compiler/scanner"
    tok "github.com/sam-caldwell/ami/src/ami/compiler/token"
    "regexp"
)

type Parser struct {
    s *scan.Scanner
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
        // placeholder: consume tokens into Bad nodes
        f.Stmts = append(f.Stmts, astpkg.Bad{Tok: p.cur})
        p.next()
    }
    return f
}


// ExtractImports finds import paths in a minimal Go-like syntax:
//   import "path"
//   import (
//     "a"
//     "b"
//   )
func ExtractImports(src string) []string {
    reSingle := regexp.MustCompile(`(?m)\bimport\s+"([^"]+)"`)
    reBlock := regexp.MustCompile(`(?m)\bimport\s*\((?s:.*?)\)`)
    imports := []string{}
    for _, m := range reSingle.FindAllStringSubmatch(src, -1) {
        if len(m) > 1 { imports = append(imports, m[1]) }
    }
    for _, blk := range reBlock.FindAllString(src, -1) {
        reQ := regexp.MustCompile(`"([^"]+)"`)
        for _, m := range reQ.FindAllStringSubmatch(blk, -1) {
            if len(m) > 1 { imports = append(imports, m[1]) }
        }
    }
    return imports
}
