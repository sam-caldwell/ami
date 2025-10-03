package sem

import (
    "testing"
    "github.com/sam-caldwell/ami/src/ami/compiler/types"
)

func Test_splitTopAllText(t *testing.T) {
    cases := map[string][]string{
        "a,b,c": {"a","b","c"},
        "map<int,string>,set<float>": {"map<int,string>","set<float>"},
        "Struct{a:int,b:string},Union<int,string>": {"Struct{a:int,b:string}","Union<int,string>"},
        "'a,b',\"c,d\"": {"'a,b'","\"c,d\""},
    }
    for in, want := range cases {
        got := splitTopAllText(in)
        if len(got) != len(want) { t.Fatalf("%q => %v", in, got) }
    }
}

func Test_unionContains(t *testing.T) {
    u := types.Union{Alts: []types.Type{types.Primitive{K: types.Int}, types.Primitive{K: types.Bool}}}
    if !unionContains(u, types.Primitive{K: types.Int}) { t.Fatalf("int in union") }
    if unionContains(u, types.Primitive{K: types.String}) { t.Fatalf("string not in union") }
    u2 := types.Union{Alts: []types.Type{types.Primitive{K: types.Bool}}}
    if !unionContains(u, u2) { t.Fatalf("subset union") }
}

func Test_validPositiveDuration(t *testing.T) {
    good := []string{"100ms","2s","3m","1h"," 5s "}
    bad := []string{"","s","10x","-1s"}
    for _, s := range good { if !validPositiveDuration(s) { t.Fatalf("good: %q", s) } }
    for _, s := range bad { if validPositiveDuration(s) { t.Fatalf("bad: %q", s) } }
}

func Test_SetStrictDedupUnderPartition(t *testing.T) {
    SetStrictDedupUnderPartition(false)
    if StrictDedupUnderPartition { t.Fatalf("expected false") }
    SetStrictDedupUnderPartition(true)
    if !StrictDedupUnderPartition { t.Fatalf("expected true") }
    SetStrictDedupUnderPartition(false) // restore
}

func Test_paramsToResults(t *testing.T) {
    in := map[string][]string{"a": {"x","y"}}
    out := paramsToResults(in)
    if len(out) != 1 || len(out["a"]) != 2 { t.Fatalf("%v", out) }
}

func Test_prim(t *testing.T) {
    good := []string{"bool","int","uint64","float32","string","byte","rune"}
    bad := []string{"","any","map","Struct{}"}
    for _, s := range good { if !prim(s) { t.Fatalf("good prim: %s", s) } }
    for _, s := range bad { if prim(s) { t.Fatalf("bad prim: %s", s) } }
}

func Test_ioAllowedIngress_Egress(t *testing.T) {
    if !ioAllowedIngress("io.stdin") { t.Fatalf("stdin ingress") }
    if !ioAllowedIngress("io.file.read") { t.Fatalf("read ingress") }
    if ioAllowedIngress("io.file.write") { t.Fatalf("write ingress forbidden") }
    if !ioAllowedEgress("io.stdout") { t.Fatalf("stdout egress") }
    if !ioAllowedEgress("io.net.send") { t.Fatalf("net send egress") }
    if !ioAllowedEgress("io.file.create") { t.Fatalf("create egress") }
}
