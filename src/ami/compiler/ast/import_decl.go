package ast

// ImportDecl captures an import path with optional alias and version constraint.
type ImportDecl struct {
    Path       string
    Alias      string
    Constraint string
    Pos        Position
    Comments   []Comment
}

func (ImportDecl) isNode() {}

