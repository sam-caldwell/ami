package types

import "testing"

func TestEqual_AllVariants(t *testing.T) {
    // Primitive
    if !Equal(Primitive{K:Int}, Primitive{K:Int}) { t.Fatalf("prim equal") }
    if Equal(Primitive{K:Int}, Primitive{K:Bool}) { t.Fatalf("prim not equal") }
    // Named
    if !Equal(Named{Name:"T"}, Named{Name:"T"}) { t.Fatalf("named equal") }
    if Equal(Named{Name:"A"}, Named{Name:"B"}) { t.Fatalf("named not equal") }
    // Generic
    g1 := Generic{Name:"Owned", Args: []Type{Primitive{K:Int}}}
    g2 := Generic{Name:"Owned", Args: []Type{Primitive{K:Int}}}
    g3 := Generic{Name:"Owned", Args: []Type{Primitive{K:Bool}}}
    if !Equal(g1, g2) || Equal(g1, g3) { t.Fatalf("generic eq") }
    // Struct (order-insensitive by field name)
    s1 := Struct{Fields: map[string]Type{"a": Primitive{K:Int}, "b": Primitive{K:Bool}}}
    s2 := Struct{Fields: map[string]Type{"b": Primitive{K:Bool}, "a": Primitive{K:Int}}}
    if !Equal(s1, s2) { t.Fatalf("struct equal by fields") }
    // Optional
    if !Equal(Optional{Inner: Primitive{K:Int}}, Optional{Inner: Primitive{K:Int}}) { t.Fatalf("optional equal") }
    // Union (order-insensitive)
    u1 := Union{Alts: []Type{Primitive{K:Int}, Primitive{K:Bool}}}
    u2 := Union{Alts: []Type{Primitive{K:Bool}, Primitive{K:Int}}}
    u3 := Union{Alts: []Type{Primitive{K:String}}}
    if !Equal(u1, u2) || Equal(u1, u3) { t.Fatalf("union eq") }
    // Default fallback: compare String() equality
    if !Equal(Slice{Elem: Primitive{K:Int}}, Slice{Elem: Primitive{K:Int}}) { t.Fatalf("slice equal") }
}

