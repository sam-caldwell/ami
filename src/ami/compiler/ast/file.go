package ast

// File represents a single source file after parsing.
type File struct {
    PackageName string
    PackagePos  source.Position
    Pragmas     []Pragma
    Decls       []Decl
}

func (*File) isNode() {}
