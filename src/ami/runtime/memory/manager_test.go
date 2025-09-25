package memory

import "testing"

func TestManager_AllocAndRelease_CountsPerDomain(t *testing.T) {
    m := NewManager()
    h1 := m.Alloc(Event, 2)
    h2 := m.Alloc(State, 3)
    h3 := m.Alloc(Ephemeral, 1)
    st := m.Stats()
    if st[Event] != 2 || st[State] != 3 || st[Ephemeral] != 1 { t.Fatalf("unexpected stats: %+v", st) }
    h1.Release()
    h1.Release() // idempotent
    h3.Release()
    st = m.Stats()
    if st[Event] != 0 || st[State] != 3 || st[Ephemeral] != 0 { t.Fatalf("unexpected stats after release: %+v", st) }
    h2.Release()
    st = m.Stats()
    if st[State] != 0 { t.Fatalf("state not zero: %+v", st) }
}

