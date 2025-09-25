package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

type File struct {
    Stmts []Node
}

type Node interface{ isNode() }

type Bad struct{ Tok tok.Token }
func (Bad) isNode() {}
