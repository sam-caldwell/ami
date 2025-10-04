package gpu

import "testing"

func TestBlocking_ConvertsPanic(t *testing.T) {
    err := Blocking(func() error { panic("boom") })
    if err == nil || err.Error() == "" { t.Fatalf("expected panic converted to error; got %v", err) }
}

