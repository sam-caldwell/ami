package ast

import "github.com/sam-caldwell/ami/src/ami/compiler/source"

// GPUBlockStmt represents a GPU-specific code block: gpu(attrs){ ...source... }.
// The Source is captured verbatim between braces; Attrs are parsed as key=value args
// using existing attribute argument parsing (flattened as Arg Text strings).
type GPUBlockStmt struct {
    Pos    source.Position
    Attrs  []Arg
    LBrace source.Position
    RBrace source.Position
    Source string
}

func (*GPUBlockStmt) isStmt() {}

