package ast

// StructDecl represents a struct declaration with named fields.
type StructDecl struct {
    Name     string
    Fields   []Field
    Pos      Position
    Comments []Comment
}

// Field is a named field with an associated type.
type Field struct {
    Name string
    Type TypeRef
}

func (StructDecl) isNode() {}

