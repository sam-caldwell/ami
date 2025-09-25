package kvstore

import "testing"

func TestStore_DeleteOnReadCount_OneTime(t *testing.T) {
    s := New(Options{})
    defer s.Close()
    s.Put("secret", []byte("xyz"), MaxReads(1))

    if v, ok := s.Get("secret"); !ok || string(v.([]byte)) != "xyz" { t.Fatalf("first get should succeed") }
    if _, ok := s.Get("secret"); ok { t.Fatalf("second get should miss after delete-on-read") }
}

