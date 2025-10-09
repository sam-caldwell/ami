package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// FuncDecl represents a function declaration (scaffold).
type FuncDecl struct {
    Pos     source.Position // position of 'func'
    NamePos source.Position
    Name    string
    // Optional method receiver (for method-style declarations):
    // func (RecvName RecvType) Name(...)
    RecvName    string
    RecvNamePos source.Position
    RecvType    string
    RecvTypePos source.Position
    TypeParams []TypeParam
    Params  []Param
    Results []Result // tuple of result types (scaffold)
    Body    *BlockStmt
    Leading []Comment
    Decorators []Decorator
    ParamsLParen  source.Position
    ParamsRParen  source.Position
    ResultsLParen source.Position
    ResultsRParen source.Position
}

func (*FuncDecl) isNode() {}
