package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// EnumDecl represents an enum declaration: enum Name { A, B, C }
type EnumDecl struct {
    Pos     source.Position
    NamePos source.Position
    Name    string
    LBrace  source.Position
    Members []EnumMember
    RBrace  source.Position
}

// EnumMember is a single enum member name and its position.
type EnumMember struct {
    Pos  source.Position
    Name string
}

func (*EnumDecl) isNode() {}

