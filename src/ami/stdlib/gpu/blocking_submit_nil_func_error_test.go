package gpu

import "testing"

func TestBlockingSubmit_NilFunc_Error(t *testing.T) {
    if err := BlockingSubmit(nil); err == nil { t.Fatalf("expected error for nil submit") }
}

