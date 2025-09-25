package root

import "testing"

func TestSplitIdents(t *testing.T) {
    cases := []struct{ in string; want []string }{
        {"alpha beta,gamma", []string{"alpha","beta","gamma"}},
        {"x1+y2-z3", []string{"x1","y2","z3"}},
        {"_ok 123 not_ident? yes_", []string{"_ok","123","not_ident","yes_"}},
        {"", nil},
    }
    for _, c := range cases {
        got := splitIdents(c.in)
        if len(got) != len(c.want) {
            t.Fatalf("len mismatch: in=%q got=%v want=%v", c.in, got, c.want)
        }
        for i := range got {
            if got[i] != c.want[i] {
                t.Fatalf("item %d mismatch: in=%q got=%v want=%v", i, c.in, got, c.want)
            }
        }
    }
}

