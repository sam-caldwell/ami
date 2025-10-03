package merge

import "testing"

func Test_less_DescAndStableAndKey(t *testing.T) {
    // Desc order
    a := item{keys: []any{2}, seq: 1}
    b := item{keys: []any{1}, seq: 2}
    p := Plan{Sort: []SortKey{{Field: "x", Order: "desc"}}}
    if !less(a, b, p) { t.Fatalf("expected a>b under desc") }
    // Stable tiebreaker
    a2 := item{keys: []any{1}, seq: 1}
    b2 := item{keys: []any{1}, seq: 2}
    p2 := Plan{Sort: []SortKey{{Field: "x", Order: "asc"}}, Stable: true}
    if !less(a2, b2, p2) { t.Fatalf("expected a2<b2 due seq under stable") }
    // Key tie-breaker
    a3 := item{keys: []any{1}, seq: 5, key: "a"}
    b3 := item{keys: []any{1}, seq: 4, key: "b"}
    p3 := Plan{Sort: []SortKey{{Field: "x", Order: "asc"}}, Key: "k"}
    if !less(a3, b3, p3) { t.Fatalf("expected key a<b tie-break") }
}

