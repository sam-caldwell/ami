package types

import "testing"

func TestComposite_String_Render(t *testing.T) {
    if got := (Pointer{Base: TString}).String(); got != "*string" {
        t.Fatalf("ptr: %q", got)
    }
    if got := (Slice{Elem: TInt}).String(); got != "[]int" {
        t.Fatalf("slice: %q", got)
    }
    if got := (Map{Key: TString, Value: TInt}).String(); got != "map<string,int>" {
        t.Fatalf("map: %q", got)
    }
    if got := (Set{Elem: TString}).String(); got != "set<string>" {
        t.Fatalf("set: %q", got)
    }
    if got := (SliceTy{Elem: TString}).String(); got != "slice<string>" {
        t.Fatalf("sliceTy: %q", got)
    }
}
