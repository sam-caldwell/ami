package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// FuncDecl represents a function declaration with an empty body scaffold.
type FuncDecl struct {
    Pos  source.Position // position of 'func'
    Name string
}

func (*FuncDecl) isNode() {}

