package exit

import "testing"

// TestError_StructFields validates the Error struct holds
// the provided Code and Msg values.
func TestError_StructFields(t *testing.T) {
    e := Error{Code: IO, Msg: "io failure"}
    if e.Code != IO {
        t.Fatalf("Code=%v, want %v", e.Code, IO)
    }
    if e.Msg != "io failure" {
        t.Fatalf("Msg=%q, want %q", e.Msg, "io failure")
    }
}
