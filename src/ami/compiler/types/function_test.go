package types

import "testing"

func TestFunction_String(t *testing.T) {
    f := Function{
        Params:  []Type{Primitive{K: Int}, Generic{Name: "Event", Args: []Type{Primitive{K: String}}}},
        Results: []Type{Primitive{K: Bool}},
    }
    if got := f.String(); got != "func(int,Event<string>) -> (bool)" {
        t.Fatalf("string: %s", got)
    }
}

