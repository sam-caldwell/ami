package bytes

import (
	stdbytes "bytes"
	"testing"
)

func TestBytes_Basics_Happy(t *testing.T) {
	if !Contains([]byte("hello"), []byte("ell")) {
		t.Fatal("Contains failed")
	}
	if Compare([]byte("a"), []byte("b")) >= 0 {
		t.Fatal("Compare failed")
	}
	if Index([]byte("banana"), []byte("na")) != 2 {
		t.Fatalf("Index expected 2")
	}
	if LastIndex([]byte("banana"), []byte("na")) != 4 {
		t.Fatalf("LastIndex expected 4")
	}
	got := Join([][]byte{[]byte("a"), []byte("b"), []byte("c")}, []byte{','})
	if string(got) != "a,b,c" {
		t.Fatalf("Join got %q", string(got))
	}
	parts := Split([]byte("a|b||c"), []byte("|"))
	if len(parts) != 4 {
		t.Fatalf("Split len %d", len(parts))
	}
	rep := Replace([]byte("a-b-b-b"), []byte("b"), []byte("x"), 2)
	if string(rep) != "a-x-x-b" {
		t.Fatalf("Replace limited failed")
	}

	// ensure determinism: Join + Split round trip
	if got2 := stdbytes.Join(parts, []byte("|")); string(got2) != "a|b||c" {
		t.Fatalf("round trip failed: %q", string(got2))
	}
}

func TestBytes_Sad(t *testing.T) {
	if Contains([]byte("hello"), []byte("zzz")) {
		t.Fatal("Contains false positive")
	}
	if Index([]byte("abc"), []byte("zzz")) != -1 {
		t.Fatal("Index expected -1")
	}
	if LastIndex([]byte("abc"), []byte("zzz")) != -1 {
		t.Fatal("LastIndex expected -1")
	}
}
