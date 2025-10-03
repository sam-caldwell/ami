package merge

import "testing"

func Test_ftoa(t *testing.T) {
    if ftoa(3.14) == "" || ftoa(1.0) == "" { t.Fatalf("ftoa empty") }
}

