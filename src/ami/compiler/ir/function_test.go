package ir

import "testing"

func TestFunction_Struct(t *testing.T) {
    fn := Function{Name: "F", Params: []Value{{ID: "p0"}}, Results: []Value{{ID: "r0"}}}
    if fn.Name != "F" || len(fn.Params) != 1 || len(fn.Results) != 1 { t.Fatalf("unexpected: %+v", fn) }
}

