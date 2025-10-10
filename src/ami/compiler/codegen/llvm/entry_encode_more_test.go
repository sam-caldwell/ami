package llvm

import "testing"

func Test_encodeCString_EscapesAndTerminates(t *testing.T) {
    // Contains quote, backslash, control (0x01) and 0xFF which must be escaped
    in := "A\"\\B\x01\xFFZ"
    got := encodeCString(in)
    // Expect: A\22 \5C B \001 \377 Z and final \00 terminator
    if want := "A\\22\\5CB\\001\\377Z\\00"; got != want {
        t.Fatalf("encodeCString mismatch:\n got: %q\nwant: %q", got, want)
    }
}
