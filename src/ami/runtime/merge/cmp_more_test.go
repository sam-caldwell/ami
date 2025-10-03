package merge

import (
    "testing"
    "time"
)

func Test_cmp(t *testing.T) {
    if cmp(false, true) != -1 || cmp(true, false) != 1 || cmp(true, true) != 0 { t.Fatalf("bool cmp") }
    if cmp(1, 2) != -1 || cmp(2, 1) != 1 || cmp(1, 1) != 0 { t.Fatalf("int cmp") }
    if cmp(int64(1), int64(2)) != -1 || cmp(int64(2), int64(1)) != 1 || cmp(int64(1), int64(1)) != 0 { t.Fatalf("int64 cmp") }
    if cmp(int64(1), 2) != -1 || cmp(int64(2), 1) != 1 || cmp(int64(1), 1) != 0 { t.Fatalf("mixed int64/int cmp") }
    if cmp(1.0, 2.0) != -1 || cmp(2.0, 1.0) != 1 || cmp(1.0, 1.0) != 0 { t.Fatalf("float cmp") }
    if cmp("a", "b") != -1 || cmp("b", "a") != 1 || cmp("a", "a") != 0 { t.Fatalf("string cmp") }
    t1 := time.Unix(1,0); t2 := time.Unix(2,0)
    if cmp(t1, t2) != -1 || cmp(t2, t1) != 1 || cmp(t1, t1) != 0 { t.Fatalf("time cmp") }
    if cmp(struct{}{}, struct{}{}) != 0 { t.Fatalf("default cmp 0") }
}

