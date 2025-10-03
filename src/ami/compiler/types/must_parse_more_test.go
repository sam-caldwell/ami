package types

import "testing"

func TestMustParse_Basic(t *testing.T) {
    cases := []string{
        "int",
        "Struct{a:int,b:string}",
        "Union<int,string>",
        "slice<int>",
        "set<string>",
        "map<int,string>",
        "Owned<int>",
    }
    for _, s := range cases {
        if MustParse(s).String() == "" { t.Fatalf("%s", s) }
    }
}

