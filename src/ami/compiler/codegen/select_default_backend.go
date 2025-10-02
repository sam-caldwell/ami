package codegen

import "fmt"

// SelectDefaultBackend sets the process-wide default backend by name.
func SelectDefaultBackend(name string) error {
    b, ok := FromName(name)
    if !ok { return fmt.Errorf("unknown backend: %s", name) }
    defaultBackend = b
    return nil
}

