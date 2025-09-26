package ast

// PackageDecl represents a package declaration.
type PackageDecl struct{ Name string }

func (PackageDecl) isNode() {}

