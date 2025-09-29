package kvstore

import (
    "sync"
    "testing"
)

func TestStore_Concurrency_PutGetConsistent(t *testing.T) {
    s := New()
    s.SetCapacity(0) // no eviction
    var wg sync.WaitGroup
    writers := 8
    per := 50
    wg.Add(writers)
    for i := 0; i < writers; i++ {
        i := i
        go func() {
            defer wg.Done()
            for j := 0; j < per; j++ {
                s.Put(key(i, j), i*1000+j)
                if v, ok := s.Get(key(i, j)); !ok || v.(int) != i*1000+j {
                    t.Errorf("get mismatch for %s", key(i, j))
                }
            }
        }()
    }
    wg.Wait()
    // spot check a few entries
    for i := 0; i < writers; i++ {
        if v, ok := s.Get(key(i, per-1)); !ok || v.(int) != i*1000+(per-1) {
            t.Fatalf("final get mismatch: %v %v", v, ok)
        }
    }
}

func key(i, j int) string { return "k_" + itoa(i) + "_" + itoa(j) }

func itoa(n int) string {
    // simple, allocation-free integer to string for small positive ints
    if n == 0 { return "0" }
    buf := make([]byte, 0, 8)
    var tmp [8]byte
    p := len(tmp)
    for n > 0 {
        p--
        tmp[p] = byte('0' + n%10)
        n /= 10
    }
    buf = append(buf, tmp[p:]...)
    return string(buf)
}

