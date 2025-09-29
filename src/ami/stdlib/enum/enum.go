package enum

import (
    "encoding/json"
    "fmt"
)

// Descriptor describes a generated enum: its type name and canonical member names
// in ordinal order (0..N-1). Names are case-sensitive and must be unique.
type Descriptor struct {
    Name  string
    Names []string
    idx   map[string]int
}

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

// MustNewDescriptor is like NewDescriptor but panics on error (for generated code).
func MustNewDescriptor(typeName string, names []string) Descriptor {
    d, err := NewDescriptor(typeName, names)
    if err != nil { panic(err) }
    return d
}

// String returns the canonical name for ordinal v, or "" if invalid.
func String(d Descriptor, v int) string {
    if v >= 0 && v < len(d.Names) { return d.Names[v] }
    return ""
}

// Ordinal returns the zero-based ordinal for v (identity in this model).
func Ordinal(v int) int { return v }

// IsValid reports whether v is within the enum range.
func IsValid(d Descriptor, v int) bool { return v >= 0 && v < len(d.Names) }

// Values returns all enum ordinals in canonical order.
func Values(d Descriptor) []int {
    out := make([]int, len(d.Names))
    for i := range out { out[i] = i }
    return out
}

// Names returns all canonical enum names in canonical order.
func Names(d Descriptor) []string { return append([]string(nil), d.Names...) }

// Parse maps a canonical name to its ordinal.
func Parse(d Descriptor, s string) (int, error) {
    if i, ok := d.idx[s]; ok { return i, nil }
    return -1, fmt.Errorf("E_ENUM_PARSE: unknown name %q for %s", s, d.Name)
}

// MustParse panics if name is invalid.
func MustParse(d Descriptor, s string) int {
    v, err := Parse(d, s)
    if err != nil { panic(err) }
    return v
}

// FromOrdinal validates i and returns it when valid.
func FromOrdinal(d Descriptor, i int) (int, error) {
    if IsValid(d, i) { return i, nil }
    return -1, fmt.Errorf("E_ENUM_ORDINAL: invalid ordinal %d for %s", i, d.Name)
}

// Value wraps an enum value with its descriptor to provide JSON/Text methods.
type Value struct {
    D Descriptor
    V int
}

func (v Value) String() string { return String(v.D, v.V) }
func (v Value) GoString() string {
    name := String(v.D, v.V)
    if name == "" { name = "<invalid>" }
    return fmt.Sprintf("%s(%s)", v.D.Name, name)
}

func (v Value) MarshalJSON() ([]byte, error) {
    name := String(v.D, v.V)
    if name == "" { return nil, fmt.Errorf("E_ENUM_ORDINAL: invalid ordinal %d for %s", v.V, v.D.Name) }
    return json.Marshal(name)
}

func (v *Value) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil { return err }
    ord, err := Parse(v.D, s)
    if err != nil { return err }
    v.V = ord
    return nil
}

func (v Value) MarshalText() ([]byte, error) {
    name := String(v.D, v.V)
    if name == "" { return nil, fmt.Errorf("E_ENUM_ORDINAL: invalid ordinal %d for %s", v.V, v.D.Name) }
    return []byte(name), nil
}

func (v *Value) UnmarshalText(b []byte) error {
    ord, err := Parse(v.D, string(b))
    if err != nil { return err }
    v.V = ord
    return nil
}

