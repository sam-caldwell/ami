package types

import (
    "sort"
    "strings"
)

// Struct represents a simple record/object type with named fields.
type Struct struct{
    Fields map[string]Type
}

func (s Struct) String() string {
    if len(s.Fields) == 0 { return "Struct{}" }
    keys := make([]string, 0, len(s.Fields))
    for k := range s.Fields { keys = append(keys, k) }
    sort.Strings(keys)
    parts := make([]string, 0, len(keys))
    for _, k := range keys { parts = append(parts, k+":"+s.Fields[k].String()) }
    return "Struct{" + strings.Join(parts, ",") + "}"
}

