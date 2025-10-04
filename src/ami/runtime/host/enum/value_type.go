package enum

import (
    "encoding/json"
    "fmt"
)

// Value wraps an enum value with its descriptor to provide JSON/Text methods.
type Value struct {
    D Descriptor
    V int
}

func (v *Value) String() string { return String(v.D, v.V) }

func (v *Value) GoString() string {
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

func (v *Value) MarshalText() ([]byte, error) {
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

