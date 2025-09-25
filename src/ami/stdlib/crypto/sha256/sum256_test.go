package sha256

import (
	"encoding/hex"
	"testing"
)

func TestSHA256_Sum256_Vectors(t *testing.T) {
	tests := []struct{ in, hexOut string }{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
	}
	for _, tc := range tests {
		got := Sum256([]byte(tc.in))
		if hex.EncodeToString(got[:]) != tc.hexOut {
			t.Fatalf("Sum256(%q) mismatch", tc.in)
		}
	}
}

func TestSHA256_Hasher_StreamVsOneShot(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")
	// one-shot
	one := Sum256(data)

	// streaming in chunks
	h := New()
	for _, ch := range [][]byte{data[:10], data[10:20], data[20:]} {
		if _, err := h.Write(ch); err != nil {
			t.Fatal(err)
		}
	}
	two := h.Sum()
	if one != two {
		t.Fatal("streaming hash mismatch")
	}

	// Reset and write again
	h.Reset()
	if _, err := h.Write(data); err != nil {
		t.Fatal(err)
	}
	if h.Sum() != one {
		t.Fatal("reset+write mismatch")
	}
}
