package exit

import "testing"

func TestCode_Int(t *testing.T) {
    cases := []struct {
        name string
        code Code
        want int
    }{
        {"OK", OK, 0},
        {"Internal", Internal, 1},
        {"User", User, 2},
        {"IO", IO, 3},
        {"Integrity", Integrity, 4},
        {"NetworkAlias", Network, int(Integrity)},
    }
    for _, tc := range cases {
        if got := tc.code.Int(); got != tc.want {
            t.Fatalf("%s.Int()=%d, want %d", tc.name, got, tc.want)
        }
    }

    // Explicitly assert alias equivalence for clarity/stability.
    if Network != Integrity {
        t.Fatalf("Network(%d) != Integrity(%d)", Network, Integrity)
    }
}
