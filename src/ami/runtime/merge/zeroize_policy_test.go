package merge

import (
    "testing"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "time"
)

// helper zeroizer: zeroizes any []byte found at top-level in a map payload; returns bytes zeroized count.
func testZeroizer(p any) int {
    n := 0
    if m, ok := p.(map[string]any); ok {
        for k, v := range m {
            if b, ok2 := v.([]byte); ok2 {
                for i := range b { b[i] = 0 }
                m[k] = b
                n += len(b)
            }
        }
    }
    return n
}

func TestZeroization_OnDropNewest(t *testing.T) {
    // Install zeroizer
    SetSensitiveZeroizer(testZeroizer)
    defer SetSensitiveZeroizer(nil)
    p := Plan{}
    p.Buffer.Capacity = 1
    p.Buffer.Policy = "dropNewest"
    op := NewOperator(p)
    a := map[string]any{"secret": []byte{1,2,3}}
    b := map[string]any{"secret": []byte{4,5,6,7}}
    if err := op.Push(ev.Event{Payload: a}); err != nil { t.Fatalf("push a: %v", err) }
    if err := op.Push(ev.Event{Payload: b}); err != nil { t.Fatalf("push b: %v", err) }
    // b should have been dropped and zeroized
    bb := b["secret"].([]byte)
    for i := range bb { if bb[i] != 0 { t.Fatalf("expected zeroized b; got %v", bb) } }
    // a remains in buffer, not zeroized
    ab := a["secret"].([]byte)
    if ab[0] == 0 && len(ab) > 0 { t.Fatalf("unexpected zeroization of retained payload: %v", ab) }
}

func TestZeroization_OnDropOldest(t *testing.T) {
    SetSensitiveZeroizer(testZeroizer)
    defer SetSensitiveZeroizer(nil)
    p := Plan{}
    p.Buffer.Capacity = 1
    p.Buffer.Policy = "dropOldest"
    op := NewOperator(p)
    a := map[string]any{"secret": []byte{9,8,7}}
    b := map[string]any{"secret": []byte{6,5,4}}
    if err := op.Push(ev.Event{Payload: a}); err != nil { t.Fatalf("push a: %v", err) }
    if err := op.Push(ev.Event{Payload: b}); err != nil { t.Fatalf("push b: %v", err) }
    // a should be dropped & zeroized, b retained
    ab := a["secret"].([]byte)
    for i := range ab { if ab[i] != 0 { t.Fatalf("expected zeroized a; got %v", ab) } }
    bb := b["secret"].([]byte)
    if bb[0] == 0 && len(bb) > 0 { t.Fatalf("unexpected zeroization of retained payload: %v", bb) }
}

func TestZeroization_OnExpireStale(t *testing.T) {
    SetSensitiveZeroizer(testZeroizer)
    defer SetSensitiveZeroizer(nil)
    p := Plan{}
    p.TimeoutMs = 10
    op := NewOperator(p)
    a := map[string]any{"secret": []byte{3,3,3}}
    b := map[string]any{"secret": []byte{4,4,4,4}}
    if err := op.Push(ev.Event{Payload: a}); err != nil { t.Fatalf("push a: %v", err) }
    if err := op.Push(ev.Event{Payload: b}); err != nil { t.Fatalf("push b: %v", err) }
    // Force expire by using future cutoff
    _ = op.ExpireStale(time.Now().Add(1 * time.Hour))
    ab := a["secret"].([]byte)
    bb := b["secret"].([]byte)
    for i := range ab { if ab[i] != 0 { t.Fatalf("expected zeroized a; got %v", ab) } }
    for i := range bb { if bb[i] != 0 { t.Fatalf("expected zeroized b; got %v", bb) } }
}

