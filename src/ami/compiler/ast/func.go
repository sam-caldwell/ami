package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// FuncDecl represents a function declaration (scaffold).
type FuncDecl struct {
    Pos     source.Position // position of 'func'
    Name    string
    Params  []Param
    Results []Result // tuple of result types (scaffold)
    Body    *BlockStmt
    Leading []Comment
}

func (*FuncDecl) isNode() {}
