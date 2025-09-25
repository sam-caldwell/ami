package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

// File represents a parsed source file.
// Minimal structure for package and import declarations.
type File struct {
    Package string
    Imports []string
    Stmts   []Node
}

type Node interface{ isNode() }

type Bad struct{ Tok tok.Token }
func (Bad) isNode() {}
