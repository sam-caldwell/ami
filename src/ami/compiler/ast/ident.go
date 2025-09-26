package ast

// Ident is a simple identifier expression with a name.
type Ident struct {
    Name string
    Pos  Position
}

func (Ident) isExpr() {}

