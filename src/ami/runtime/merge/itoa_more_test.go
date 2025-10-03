package merge

import "testing"

func Test_itoa(t *testing.T) {
    if itoa(0) != "0" || itoa(-1) != "-1" || itoa(42) != "42" { t.Fatalf("itoa bad") }
}

