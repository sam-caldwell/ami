package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// FuncDecl represents a function declaration (scaffold).
type FuncDecl struct {
    Pos     source.Position // position of 'func'
    NamePos source.Position
    Name    string
    Params  []Param
    Results []Result // tuple of result types (scaffold)
    Body    *BlockStmt
    Leading []Comment
    ParamsLParen  source.Position
    ParamsRParen  source.Position
    ResultsLParen source.Position
    ResultsRParen source.Position
}

func (*FuncDecl) isNode() {}
