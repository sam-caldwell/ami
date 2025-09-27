package sem

import "testing"

func TestTypesCompatible_Basics(t *testing.T) {
    if !typesCompatible("int", "int") { t.Fatalf("equal types") }
    if !typesCompatible("any", "int") { t.Fatalf("any wildcard") }
    if !typesCompatible("Event<T>", "Event<int>") { t.Fatalf("event generic") }
    if typesCompatible("Event<int>", "Event<string>") { t.Fatalf("event mismatch") }
    if !typesCompatible("slice<T>", "slice<int>") { t.Fatalf("slice generic") }
    if !typesCompatible("map<any,string>", "map<int,string>") { t.Fatalf("map partial any") }
}

