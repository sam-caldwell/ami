package gpu

import "testing"

func TestBlocking_NilFunc_Error(t *testing.T) {
    if err := Blocking(nil); err == nil { t.Fatalf("expected error for nil function") }
}

