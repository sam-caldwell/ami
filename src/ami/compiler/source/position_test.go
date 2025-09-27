package source

import "testing"

func TestPosition_ZeroAndFields(t *testing.T) {
    var zero Position
    if zero.Line != 0 || zero.Column != 0 || zero.Offset != 0 {
        t.Fatalf("zero position should be all zeros: %+v", zero)
    }
    p := Position{Line: 2, Column: 3, Offset: 5}
    if p.Line != 2 || p.Column != 3 || p.Offset != 5 {
        t.Fatalf("unexpected fields: %+v", p)
    }
}

