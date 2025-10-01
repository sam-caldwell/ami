package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// ParseFileCollect parses a file and returns the file and any collected errors.
func (p *Parser) ParseFileCollect() (*ast.File, []error) {
    f, _ := p.ParseFile()
    return f, p.errors
}

