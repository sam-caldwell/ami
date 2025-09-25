package strings

import "testing"

func TestStrings_Basics_Happy(t *testing.T) {
	if !Contains("hello", "ell") {
		t.Fatal("Contains failed")
	}
	if HasPrefix("hello", "he") == false {
		t.Fatal("HasPrefix failed")
	}
	if HasSuffix("hello", "lo") == false {
		t.Fatal("HasSuffix failed")
	}
	if Index("banana", "na") != 2 {
		t.Fatalf("Index expected 2")
	}
	if LastIndex("banana", "na") != 4 {
		t.Fatalf("LastIndex expected 4")
	}
	if EqualFold("Ångström", "ångström") == false {
		t.Fatal("EqualFold unicode failed")
	}

	if got := Join([]string{"a", "b", "c"}, ","); got != "a,b,c" {
		t.Fatalf("Join got %q", got)
	}
	parts := Split("a|b||c", "|")
	if len(parts) != 4 {
		t.Fatalf("Split len %d", len(parts))
	}
	if Replace("a-b-b-b", "b", "x", 2) != "a-x-x-b" {
		t.Fatalf("Replace limited failed")
	}
	if Trim("--hi--", "-") != "hi" {
		t.Fatalf("Trim failed")
	}
	if TrimSpace("  hi \n") != "hi" {
		t.Fatalf("TrimSpace failed")
	}
	if ToLower("ÄÖÜ") != "äöü" {
		t.Fatalf("ToLower failed")
	}
	if ToUpper("äöü") != "ÄÖÜ" {
		t.Fatalf("ToUpper failed")
	}
	f := Fields(" a  b\t c\n ")
	if len(f) != 3 || f[0] != "a" || f[2] != "c" {
		t.Fatalf("Fields wrong: %#v", f)
	}
}

func TestStrings_Sad(t *testing.T) {
	if Contains("hello", "zzz") {
		t.Fatal("Contains false positive")
	}
	if HasPrefix("hello", "zz") {
		t.Fatal("HasPrefix false positive")
	}
	if HasSuffix("hello", "zz") {
		t.Fatal("HasSuffix false positive")
	}
	if Index("abc", "zzz") != -1 {
		t.Fatal("Index expected -1")
	}
	if LastIndex("abc", "zzz") != -1 {
		t.Fatal("LastIndex expected -1")
	}
}
