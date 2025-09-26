package ast

// EnumDecl represents an enum declaration with named members.
type EnumDecl struct {
    Name     string
    Members  []EnumMember
    Pos      Position
    Comments []Comment
}

func (EnumDecl) isNode() {}
