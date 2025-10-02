package types

import "testing"

func testParseOptionalUnion_StringRoundtrip(t *testing.T) {
	ty, err := Parse("Optional<int>")
	if err != nil {
		t.Fatalf("parse optional: %v", err)
	}
	if ty.String() != "Optional<int>" {
		t.Fatalf("roundtrip optional: %s", ty.String())
	}
	tu, err := Parse("Union<int,string>")
	if err != nil {
		t.Fatalf("parse union: %v", err)
	}
	if tu.String() != "Union<int,string>" && tu.String() != "Union<string,int>" {
		t.Fatalf("roundtrip union: %s", tu.String())
	}
}

func testEqual_Structural_UnionOrderInsensitive(t *testing.T) {
	a, _ := Parse("Union<int,string>")
	b, _ := Parse("Union<string,int>")
	if !Equal(a, b) {
		t.Fatalf("expected union equality order-insensitive")
	}
}

func testResolveField_Deep_WithOptionalUnion(t *testing.T) {
	// Event<Struct{x:Optional<Struct{y:Union<int,string>}>}>
	root, err := Parse("Event<Struct{x:Optional<Struct{y:Union<int,string>}>}>")
	if err != nil {
		t.Fatalf("parse root: %v", err)
	}
	leaf, ok := ResolveField(root, "x.y")
	if !ok {
		t.Fatalf("failed to resolve x.y")
	}
	// Expect Optional<Union<int,string>>
	want, _ := Parse("Optional<Union<int,string>>")
	if !Equal(leaf, want) {
		t.Fatalf("got %s want %s", leaf.String(), want.String())
	}
	if !IsOrderable(leaf) {
		t.Fatalf("expected Optional<Union<int,string>> to be orderable")
	}
}

func testResolveField_UnorderableLeaf(t *testing.T) {
	root, _ := Parse("Event<Struct{k:Struct{a:int}}>")
	leaf, ok := ResolveField(root, "k")
	if !ok {
		t.Fatalf("resolve k")
	}
	if IsOrderable(leaf) {
		t.Fatalf("struct leaf should be unorderable")
	}
}
