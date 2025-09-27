package ir

// Module groups functions for a package.
type Module struct {
    Package   string
    Functions []Function
    Directives []Directive
}
