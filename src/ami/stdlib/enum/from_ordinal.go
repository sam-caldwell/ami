package enum

import "fmt"

// FromOrdinal validates i and returns it when valid.
func FromOrdinal(d Descriptor, i int) (int, error) {
    if IsValid(d, i) { return i, nil }
    return -1, fmt.Errorf("E_ENUM_ORDINAL: invalid ordinal %d for %s", i, d.Name)
}

