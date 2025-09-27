package token

import "testing"

// TestKind_String_AllEnumerated iterates over the full enumerated range
// of token kinds and ensures String() returns a non-empty, non-"Unknown"
// label for every defined kind except the explicit Unknown sentinel.
func TestKind_String_AllEnumerated(t *testing.T) {
    // Iterate from the first defined kind through the last (BlockComment).
    for i := int(Unknown); i <= int(BlockComment); i++ {
        k := Kind(i)
        s := k.String()
        if k == Unknown {
            if s != "Unknown" {
                t.Fatalf("Unknown.String() => %q; want Unknown", s)
            }
            continue
        }
        if s == "" || s == "Unknown" {
            t.Fatalf("Kind(%d).String() not mapped (got %q)", i, s)
        }
    }
}

