package exit

import "testing"

func TestCode_ConstantsAndIntValues(t *testing.T) {
    // Verify mapping of constants to ints is stable
    if OK.Int() != 0 { t.Fatalf("OK.Int()=%d", OK.Int()) }
    if Internal.Int() != 1 { t.Fatalf("Internal.Int()=%d", Internal.Int()) }
    if User.Int() != 2 { t.Fatalf("User.Int()=%d", User.Int()) }
    if IO.Int() != 3 { t.Fatalf("IO.Int()=%d", IO.Int()) }
    if Integrity.Int() != 4 { t.Fatalf("Integrity.Int()=%d", Integrity.Int()) }
}

