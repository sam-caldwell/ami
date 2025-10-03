package tester

import "testing"

func Test_asInt_Conversions(t *testing.T) {
    cases := []any{int(3), int32(4), int64(5), float64(6.5), float32(7.2)}
    for _, v := range cases {
        if n, ok := asInt(v); !ok || n == 0 { t.Fatalf("convert %T failed: %v %v", v, n, ok) }
    }
    if _, ok := asInt("bad"); ok { t.Fatalf("expected false for non-number") }
}

