package rand

import "testing"

func TestRand_DeterministicSequences(t *testing.T) {
	p1 := New(42)
	p2 := New(42)
	for i := 0; i < 5; i++ {
		if p1.Intn(1000) != p2.Intn(1000) {
			t.Fatal("Intn sequences differ")
		}
	}
	if p1.Uint64() != p2.Uint64() {
		t.Fatal("Uint64 sequences differ")
	}
	a := p1.Perm(5)
	b := p2.Perm(5)
	for i := range a {
		if a[i] != b[i] {
			t.Fatal("Perm sequences differ")
		}
	}
	buf1 := make([]byte, 16)
	buf2 := make([]byte, 16)
	if _, err := p1.Read(buf1); err != nil {
		t.Fatal(err)
	}
	if _, err := p2.Read(buf2); err != nil {
		t.Fatal(err)
	}
	for i := range buf1 {
		if buf1[i] != buf2[i] {
			t.Fatal("Read bytes differ")
		}
	}
}

func TestRand_Sad_IntnPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for Intn(0)")
		}
	}()
	New(1).Intn(0)
}
