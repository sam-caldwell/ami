package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// Decorator represents a function decorator like @name or @name(arg1, ...).
// It is attached to FuncDecl and preserves source order.
type Decorator struct {
    Pos     source.Position // position of '@'
    NamePos source.Position
    Name    string
    LParen  source.Position // zero if no args
    Args    []Expr          // args are simple literals or identifiers
    RParen  source.Position // zero if no args
}

func (*Decorator) isNode() {}

