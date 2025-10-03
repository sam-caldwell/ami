package merge

import "testing"

func Test_toKey(t *testing.T) {
    if toKey("x") != "x" { t.Fatalf("string") }
    if toKey(3) != "3" { t.Fatalf("int") }
    if toKey(int64(3)) != "3" { t.Fatalf("int64") }
    if toKey(3.5) == "" { t.Fatalf("float64") }
    if toKey(true) != "true" || toKey(false) != "false" { t.Fatalf("bool") }
    if toKey(struct{}{}) != "" { t.Fatalf("default should be empty") }
}

