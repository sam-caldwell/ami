package ast

// EnumDecl represents an enum declaration with named members.
type EnumDecl struct {
    Name     string
    Members  []EnumMember
    Pos      Position
    Comments []Comment
}

// EnumMember is a single member in an enum with an optional literal value.
type EnumMember struct {
    Name  string
    Value string
}

func (EnumDecl) isNode() {}

