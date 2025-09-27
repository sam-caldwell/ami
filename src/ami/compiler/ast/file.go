package ast

// File represents a single source file after parsing.
type File struct {
    PackageName string
    Decls       []Decl
}

func (*File) isNode() {}

