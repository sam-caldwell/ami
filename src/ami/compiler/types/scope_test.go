package types

import "testing"

func TestScope_InsertLookup_Order(t *testing.T) {
    s := NewScope(nil)
    if err := s.Insert(&Object{Kind: ObjFunc, Name: "b", Type: TString}); err != nil { t.Fatalf("insert b: %v", err) }
    if err := s.Insert(&Object{Kind: ObjFunc, Name: "a", Type: TString}); err != nil { t.Fatalf("insert a: %v", err) }
    if err := s.Insert(&Object{Kind: ObjType, Name: "T", Type: TBool}); err != nil { t.Fatalf("insert T: %v", err) }
    if got, want := s.Lookup("a").Name, "a"; got != want { t.Fatalf("lookup a=%q", got) }
    if got := s.Lookup("z"); got != nil { t.Fatalf("lookup z expected nil") }
    names := s.Names()
    want := []string{"b", "a", "T"}
    if len(names) != len(want) { t.Fatalf("names len=%d want %d", len(names), len(want)) }
    for i := range want { if names[i] != want[i] { t.Fatalf("names[%d]=%q want %q", i, names[i], want[i]) } }
}

func TestScope_Duplicate_Error(t *testing.T) {
    s := NewScope(nil)
    if err := s.Insert(&Object{Kind: ObjFunc, Name: "x", Type: TInt}); err != nil { t.Fatalf("insert: %v", err) }
    if err := s.Insert(&Object{Kind: ObjFunc, Name: "x", Type: TInt}); err == nil { t.Fatalf("expected duplicate error") }
}

