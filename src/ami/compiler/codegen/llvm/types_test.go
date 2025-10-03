package llvm

import "testing"

func Test_mapType_CoversBasicAndContainers(t *testing.T) {
    cases := map[string]string{
        "": "void",
        "void": "void",
        "bool": "i1",
        "int8": "i8", "int16": "i16", "int32": "i32", "int64": "i64", "int": "i64",
        "uint8": "i8", "uint16": "i16", "uint32": "i32", "uint64": "i64", "uint": "i64",
        "real": "double", "float64": "double",
        "string": "ptr",
        "Owned": "ptr",
        "Event<int>": "ptr",
        "Map<string,int>": "ptr",
    }
    for in, want := range cases {
        if got := mapType(in); got != want {
            t.Fatalf("mapType(%q)=%q want %q", in, got, want)
        }
    }
}
