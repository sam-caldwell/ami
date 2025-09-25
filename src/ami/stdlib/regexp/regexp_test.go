package regexp

import "testing"

func TestRegexp_Compile_MatchString_Happy(t *testing.T) {
	r, err := Compile(`^a.*b$`)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	if !r.MatchString("axxxb") {
		t.Fatal("expected match")
	}
	if r.MatchString("abx") {
		t.Fatal("unexpected match")
	}
}

func TestRegexp_MustCompile_Happy(t *testing.T) {
	r := MustCompile(`^[0-9]+$`)
	if !r.MatchString("12345") {
		t.Fatal("expected match")
	}
}

func TestRegexp_Sad_InvalidPattern(t *testing.T) {
	if _, err := Compile("("); err == nil {
		t.Fatal("expected error")
	}
}
