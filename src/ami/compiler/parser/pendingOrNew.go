package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/ast"

// pendingOrNew is a small helper to use or initialize p.pending safely.
func pendingOrNew(_ []ast.Comment, p *Parser) []ast.Comment { return p.pending }

