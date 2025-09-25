package tester

import "testing"

func TestRunner_Execute_SkipsByDefault(t *testing.T) {
    r := New()
    out, err := r.Execute("P", []Case{{Name: "c1"}, {Name: "c2"}})
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if len(out) != 2 { t.Fatalf("want 2 results; got %d", len(out)) }
    for _, r := range out {
        if r.Status != "skip" { t.Fatalf("expected skip; got %q", r.Status) }
        if r.Error == "" { t.Fatalf("want reason for skip") }
    }
}

