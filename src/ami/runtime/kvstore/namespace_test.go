package kvstore

import "testing"

func TestNamespace_IsolatedStores(t *testing.T) {
    ResetRegistry()
    a := Namespace("p1/n1")
    b := Namespace("p1/n2")
    if a == b { t.Fatalf("expected separate stores for different namespaces") }
    a.Put("k", 1)
    if _, ok := b.Get("k"); ok { t.Fatalf("isolation violated: b can see a's key") }
}

