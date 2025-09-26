package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

// FuncLit represents an inline function literal used in expressions.
type FuncLit struct {
    Params []Param
    Result []TypeRef
    Body   []tok.Token
    Pos    Position
}

func (FuncLit) isExpr() {}

