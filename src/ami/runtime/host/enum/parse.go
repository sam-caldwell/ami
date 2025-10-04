package enum

import "fmt"

// Parse maps a canonical name to its ordinal.
func Parse(d Descriptor, s string) (int, error) {
    if i, ok := d.idx[s]; ok { return i, nil }
    return -1, fmt.Errorf("E_ENUM_PARSE: unknown name %q for %s", s, d.Name)
}

