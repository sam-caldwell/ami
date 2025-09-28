package codegen

import "fmt"

// Backend registry and default selection.

var registry = map[string]Backend{
    "llvm": &llvmBackend{},
}

// FromName returns a backend by name.
func FromName(name string) (Backend, bool) {
    if name == "" { return defaultBackend, true }
    b, ok := registry[name]
    return b, ok
}

// SelectDefaultBackend sets the process-wide default backend by name.
func SelectDefaultBackend(name string) error {
    b, ok := FromName(name)
    if !ok { return fmt.Errorf("unknown backend: %s", name) }
    defaultBackend = b
    return nil
}

