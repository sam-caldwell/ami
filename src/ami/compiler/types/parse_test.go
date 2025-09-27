package types

import "testing"

func TestParse_Primitives(t *testing.T) {
    cases := map[string]string{
        "bool":    "bool",
        "int":     "int",
        "int64":   "int64",
        "float64": "float64",
        "string":  "string",
    }
    for in, want := range cases {
        tt, err := Parse(in)
        if err != nil { t.Fatalf("parse %q: %v", in, err) }
        if tt.String() != want { t.Fatalf("%s -> %s", in, tt.String()) }
    }
}

func TestParse_Generics_SingleArg(t *testing.T) {
    cases := map[string]string{
        "slice<int>":  "slice<int>",
        "set<string>": "set<string>",
        "Event<int>":  "Event<int>",
        "Error<string>": "Error<string>",
    }
    for in, want := range cases {
        tt, err := Parse(in)
        if err != nil { t.Fatalf("parse %q: %v", in, err) }
        if tt.String() != want { t.Fatalf("%s -> %s", in, tt.String()) }
    }
}

func TestParse_Map_TwoArgs(t *testing.T) {
    tt, err := Parse("map<string,int64>")
    if err != nil { t.Fatalf("parse: %v", err) }
    if tt.String() != "map<string,int64>" { t.Fatalf("map render: %s", tt.String()) }
}

func TestParse_UnknownBase_AsNamedGeneric(t *testing.T) {
    tt, err := Parse("Foo")
    if err != nil { t.Fatalf("parse: %v", err) }
    if tt.String() != "Foo" { t.Fatalf("Foo render: %s", tt.String()) }
}

