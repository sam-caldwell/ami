package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// ImportDecl represents a single import declaration.
// Path holds the imported module/path in source form.
type ImportDecl struct {
    Pos  source.Position
    Path string
    Leading []Comment
    PathPos source.Position
    // Alias is an optional short name used to reference the import in code.
    Alias   string
    AliasPos source.Position
    // Constraint holds an optional version constraint string as written in source,
    // e.g., ">= v1.2.3". Empty when not specified.
    Constraint string
}

func (*ImportDecl) isNode() {}
