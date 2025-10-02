package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// DurationLit represents a duration literal like 300ms, 5s, 2h45m, 1.5h
type DurationLit struct {
    Pos  source.Position
    Text string
}

func (*DurationLit) isNode() {}
func (*DurationLit) isExpr() {}

