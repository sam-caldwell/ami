package ast

// TypeRef references a type name with optional generic arguments and modifiers.
type TypeRef struct {
    Name   string
    Args   []TypeRef
    Ptr    bool
    Slice  bool
    Offset int
}

