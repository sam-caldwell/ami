package llvm

import "testing"

func Test_encodeCString_EscapesAndTerminates(t *testing.T) {
    // Contains quote, backslash, control (0x01) and 0xFF which must be hex-escaped
    in := "A\"\\B\x01\xFFZ"
    got := encodeCString(in)
    // Expect: A\" \\ B \x01 \xFF Z and final \00 terminator
    if want := "A\\\"\\\\B\\x01\\xFFZ\\00"; got != want {
        t.Fatalf("encodeCString mismatch:\n got: %q\nwant: %q", got, want)
    }
}

