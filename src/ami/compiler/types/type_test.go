package types

import "testing"

func TestPrimitive_String(t *testing.T) {
    if (Primitive{K: Bool}).String() != "bool" { t.Fatalf("bool") }
    if (Primitive{K: Int}).String() != "int" { t.Fatalf("int") }
    if (Primitive{K: Float64}).String() != "float64" { t.Fatalf("float64") }
    if (Primitive{K: String}).String() != "string" { t.Fatalf("string") }
}

func TestGeneric_String(t *testing.T) {
    g := Generic{Name: "Event", Args: []Type{Primitive{K: Int}}}
    if g.String() != "Event<int>" { t.Fatalf("Event<int>: %s", g.String()) }
    h := Generic{Name: "map", Args: []Type{Primitive{K: String}, Primitive{K: Int64}}}
    if h.String() != "map<string,int64>" { t.Fatalf("map<string,int64>: %s", h.String()) }
}
