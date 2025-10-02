package enum

// MustNewDescriptor is like NewDescriptor but panics on error (for generated code).
func MustNewDescriptor(typeName string, names []string) Descriptor {
    d, err := NewDescriptor(typeName, names)
    if err != nil { panic(err) }
    return d
}

