package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Expr is implemented by all expression nodes.
type Expr interface{ isExpr() }

type IdentExpr struct {
    Pos  source.Position
    Name string
}
func (*IdentExpr) isNode() {}
func (*IdentExpr) isExpr() {}

type StringLit struct {
    Pos   source.Position
    Value string
}
func (*StringLit) isNode() {}
func (*StringLit) isExpr() {}

type NumberLit struct {
    Pos  source.Position
    Text string
}
func (*NumberLit) isNode() {}
func (*NumberLit) isExpr() {}

// DurationLit represents a duration literal like 300ms, 5s, 2h45m, 1.5h
type DurationLit struct {
    Pos  source.Position
    Text string
}
func (*DurationLit) isNode() {}
func (*DurationLit) isExpr() {}

type CallExpr struct {
    Pos     source.Position
    Name    string
    NamePos source.Position
    LParen  source.Position
    Args    []Expr
    RParen  source.Position
}
func (*CallExpr) isNode() {}
func (*CallExpr) isExpr() {}
