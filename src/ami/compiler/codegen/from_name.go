package codegen

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

