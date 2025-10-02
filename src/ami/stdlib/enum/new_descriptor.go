package enum

import "fmt"

// NewDescriptor constructs a Descriptor and validates names.
func NewDescriptor(typeName string, names []string) (Descriptor, error) {
    if typeName == "" { typeName = "Enum" }
    d := Descriptor{Name: typeName, Names: append([]string(nil), names...)}
    d.idx = make(map[string]int, len(names))
    for i, n := range names {
        if n == "" { return Descriptor{}, fmt.Errorf("E_ENUM_PARSE: empty name at ordinal %d", i) }
        if _, dup := d.idx[n]; dup { return Descriptor{}, fmt.Errorf("E_ENUM_PARSE: duplicate name %q", n) }
        d.idx[n] = i
    }
    return d, nil
}

