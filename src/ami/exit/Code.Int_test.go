package exit

import "testing"

func TestCodeInt_FilePair(t *testing.T) {
    if OK.Int() != 0 { t.Fatalf("OK.Int() != 0") }
}

