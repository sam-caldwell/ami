package ast

// BasicLit is a basic literal like string or number.
type BasicLit struct {
    Kind  string
    Value string
    Pos   Position
}

func (BasicLit) isExpr() {}

