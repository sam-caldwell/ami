package kvstore

import (
	"sync"
	"testing"
)

func TestStore_Concurrency_PutsAndGets(t *testing.T) {
	s := New(Options{})
	defer s.Close()
	const N = 100
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Put(keyf("k", i), i)
		}()
	}
	wg.Wait()
	// Concurrent reads
	for i := 0; i < N; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, ok := s.Get(keyf("k", i)); !ok || v.(int) != i {
				t.Errorf("get mismatch for %d", i)
			}
		}()
	}
	wg.Wait()
	if len(s.Keys()) != N {
		t.Fatalf("expected %d keys; got %d", N, len(s.Keys()))
	}
}

func keyf(prefix string, n int) string { return prefix + "_" + itoa(n) }

func itoa(n int) string {
	// small helper without strconv for portability, N is small
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
