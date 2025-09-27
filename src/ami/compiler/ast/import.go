package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ImportDecl represents a single import declaration.
// Path holds the imported module/path in source form.
type ImportDecl struct {
    Pos  source.Position
    Path string
    Leading []Comment
    PathPos source.Position
}

func (*ImportDecl) isNode() {}
