package ast

// File represents a parsed source file with package, imports and declarations.
// The Stmts field mirrors Decls for legacy consumers.
type File struct {
    Package    string
    Version    string
    Imports    []string
    Decls      []Node
    Stmts      []Node
    Directives []Directive
}

