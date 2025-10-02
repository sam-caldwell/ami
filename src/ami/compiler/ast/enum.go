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

func (*EnumDecl) isNode() {}
