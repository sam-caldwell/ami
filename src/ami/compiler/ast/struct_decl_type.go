package ast

// StructDecl represents a struct declaration with named fields.
type StructDecl struct {
    Name     string
    Fields   []Field
    Pos      Position
    Comments []Comment
}

