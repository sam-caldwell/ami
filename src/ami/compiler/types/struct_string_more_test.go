package types

import "testing"

func TestStruct_String_SortedFields(t *testing.T) {
    s := Struct{Fields: map[string]Type{"b": Primitive{K:Int}, "a": Primitive{K:Bool}}}
    if s.String() != "Struct{a:bool,b:int}" { t.Fatalf("%s", s.String()) }
}

