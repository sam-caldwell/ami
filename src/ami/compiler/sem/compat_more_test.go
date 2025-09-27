package sem

import "testing"

// TestTypesCompatible_Containers exercises elemCompatible and keyVal paths.
func TestTypesCompatible_Containers(t *testing.T) {
    // slice element compatible via exact match
    if !typesCompatible("slice<int>", "slice<int>") { t.Fatal("slice exact") }
    // set with type var
    if !typesCompatible("set<T>", "set<int>") { t.Fatal("set typevar") }
    // map exact
    if !typesCompatible("map<string,int>", "map<string,int>") { t.Fatal("map exact") }
    // map with any
    if !typesCompatible("map<any,int>", "map<string,int>") { t.Fatal("map any key") }
    // mismatched
    if typesCompatible("map<string,int>", "map<int,string>") { t.Fatal("map mismatch should be false") }
}

