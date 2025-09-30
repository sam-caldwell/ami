package merge

import (
    "testing"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
)

func E(m map[string]any) ev.Event { return ev.Event{Payload: m} }

func TestMerge_Window_CapsBuffer(t *testing.T) {
    var p Plan
    p.Buffer.Capacity = 100
    p.Window = 2
    p.Sort = []SortKey{{Field:"k", Order:"asc"}}
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"k":2}))
    _ = op.Push(E(map[string]any{"k":1}))
    _ = op.Push(E(map[string]any{"k":0})) // should be ignored due to window cap
    // Pop two items, sorted within window of 2 elements
    o1, ok := op.Pop(); if !ok { t.Fatal("no o1") }
    o2, ok := op.Pop(); if !ok { t.Fatal("no o2") }
    k1 := o1.Payload.(map[string]any)["k"].(int)
    k2 := o2.Payload.(map[string]any)["k"].(int)
    if !(k1 <= k2) { t.Fatalf("unsorted within window: %d %d", k1, k2) }
    if _, ok := op.Pop(); ok { t.Fatalf("expected only 2 items due to window") }
}

func TestMerge_Watermark_DropsLate(t *testing.T) {
    now := time.Now()
    var p Plan
    p.Watermark = &Watermark{Field:"ts", LatenessMs: 100}
    op := NewOperator(p)
    late := now.Add(-1 * time.Hour)
    _ = op.Push(E(map[string]any{"ts": late.Format(time.RFC3339)}))
    if _, ok := op.Pop(); ok { t.Fatalf("expected late event to be dropped") }
}

func TestMerge_Timeout_ExpiresPartitions(t *testing.T) {
    var p Plan
    p.TimeoutMs = 10
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"k":1}))
    time.Sleep(15 * time.Millisecond)
    dropped := op.ExpireStale(time.Now())
    if dropped == 0 { t.Fatalf("expected dropped > 0, got %d", dropped) }
}

func TestMerge_Partition_RoundRobin(t *testing.T) {
    var p Plan
    p.PartitionBy = "p"
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"p":"a", "k":1}))
    _ = op.Push(E(map[string]any{"p":"b", "k":1}))
    _ = op.Push(E(map[string]any{"p":"a", "k":2}))
    // Round-robin between a and b
    o1, _ := op.Pop(); o2, _ := op.Pop(); o3, _ := op.Pop()
    p1 := o1.Payload.(map[string]any)["p"].(string)
    p2 := o2.Payload.(map[string]any)["p"].(string)
    p3 := o3.Payload.(map[string]any)["p"].(string)
    if p1 == p2 { t.Fatalf("not round-robin: %s %s %s", p1, p2, p3) }
}

func TestMerge_Tiebreakers_StableVsDefault(t *testing.T) {
    // Default: still deterministic because stable sort is used
    var p1 Plan
    p1.Sort = []SortKey{{Field:"k", Order:"asc"}}
    op1 := NewOperator(p1)
    _ = op1.Push(E(map[string]any{"k":1, "id":"a"}))
    _ = op1.Push(E(map[string]any{"k":1, "id":"b"}))
    x1,_ := op1.Pop(); x2,_ := op1.Pop()
    if x1.Payload.(map[string]any)["id"] != "a" || x2.Payload.(map[string]any)["id"] != "b" {
        t.Fatalf("default tiebreaker changed order")
    }
    // Stable: explicit
    var p2 Plan
    p2.Stable = true
    p2.Sort = []SortKey{{Field:"k", Order:"asc"}}
    op2 := NewOperator(p2)
    _ = op2.Push(E(map[string]any{"k":1, "id":"a"}))
    _ = op2.Push(E(map[string]any{"k":1, "id":"b"}))
    y1,_ := op2.Pop(); y2,_ := op2.Pop()
    if y1.Payload.(map[string]any)["id"] != "a" || y2.Payload.(map[string]any)["id"] != "b" {
        t.Fatalf("stable tiebreaker changed order")
    }
}

func TestMerge_EqualKey_Determinism_AcrossRuns(t *testing.T) {
    // Repeated runs should produce same order for equal sort keys
    var p Plan
    p.Sort = []SortKey{{Field:"k", Order:"asc"}}
    p.Stable = true
    for r := 0; r < 5; r++ {
        op := NewOperator(p)
        _ = op.Push(E(map[string]any{"k":1, "id":"a"}))
        _ = op.Push(E(map[string]any{"k":1, "id":"b"}))
        e1,_ := op.Pop(); e2,_ := op.Pop()
        if e1.Payload.(map[string]any)["id"].(string) != "a" || e2.Payload.(map[string]any)["id"].(string) != "b" {
            t.Fatalf("determinism broken on run %d: %v %v", r, e1, e2)
        }
    }
}

func TestMerge_Tiebreak_ByKey(t *testing.T) {
    var p Plan
    p.Sort = []SortKey{{Field:"k", Order:"asc"}}
    p.Key = "id"
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"k":1, "id":"a"}))
    _ = op.Push(E(map[string]any{"k":1, "id":"b"}))
    e1,_ := op.Pop(); e2,_ := op.Pop()
    if e1.Payload.(map[string]any)["id"].(string) != "a" || e2.Payload.(map[string]any)["id"].(string) != "b" {
        t.Fatalf("key tiebreak failed: %v %v", e1, e2)
    }
}

func TestMerge_Partition_SortedWithinEach(t *testing.T) {
    var p Plan
    p.PartitionBy = "p"
    p.Sort = []SortKey{{Field:"k", Order:"asc"}}
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"p":"x", "k":2}))
    _ = op.Push(E(map[string]any{"p":"x", "k":1}))
    _ = op.Push(E(map[string]any{"p":"y", "k":2}))
    _ = op.Push(E(map[string]any{"p":"y", "k":1}))
    // Collect pops until exhausted
    var xs []int
    var ys []int
    for {
        e, ok := op.Pop(); if !ok { break }
        m := e.Payload.(map[string]any)
        if m["p"].(string) == "x" { xs = append(xs, m["k"].(int)) } else { ys = append(ys, m["k"].(int)) }
    }
    if !(len(xs) == 2 && xs[0] <= xs[1]) { t.Fatalf("x not sorted: %v", xs) }
    if !(len(ys) == 2 && ys[0] <= ys[1]) { t.Fatalf("y not sorted: %v", ys) }
}

func TestMerge_Comparator_Bool_Order(t *testing.T) {
    var p Plan
    p.Sort = []SortKey{{Field:"ok", Order:"asc"}}
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"ok": true, "id": "b"}))
    _ = op.Push(E(map[string]any{"ok": false, "id": "a"}))
    e1,_ := op.Pop(); e2,_ := op.Pop()
    if e1.Payload.(map[string]any)["id"].(string) != "a" || e2.Payload.(map[string]any)["id"].(string) != "b" {
        t.Fatalf("bool comparator order wrong: %v %v", e1, e2)
    }
}

func TestMerge_Comparator_Timestamp_Order(t *testing.T) {
    var p Plan
    p.Sort = []SortKey{{Field:"ts", Order:"asc"}}
    op := NewOperator(p)
    t1 := time.Now().Add(-time.Hour)
    t2 := time.Now()
    _ = op.Push(E(map[string]any{"ts": t2, "id": "b"}))
    _ = op.Push(E(map[string]any{"ts": t1, "id": "a"}))
    e1,_ := op.Pop(); e2,_ := op.Pop()
    if e1.Payload.(map[string]any)["id"].(string) != "a" || e2.Payload.(map[string]any)["id"].(string) != "b" {
        t.Fatalf("timestamp comparator order wrong: %v %v", e1, e2)
    }
}

func TestMerge_Desc_Int_String(t *testing.T) {
    // Int desc
    var p Plan
    p.Sort = []SortKey{{Field:"k", Order:"desc"}}
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"k":1}))
    _ = op.Push(E(map[string]any{"k":2}))
    e1,_ := op.Pop(); e2,_ := op.Pop()
    if e1.Payload.(map[string]any)["k"].(int) != 2 || e2.Payload.(map[string]any)["k"].(int) != 1 {
        t.Fatalf("int desc wrong: %v %v", e1, e2)
    }
    // String desc
    var p2 Plan
    p2.Sort = []SortKey{{Field:"s", Order:"desc"}}
    op2 := NewOperator(p2)
    _ = op2.Push(E(map[string]any{"s":"a"}))
    _ = op2.Push(E(map[string]any{"s":"b"}))
    s1,_ := op2.Pop(); s2,_ := op2.Pop()
    if s1.Payload.(map[string]any)["s"].(string) != "b" || s2.Payload.(map[string]any)["s"].(string) != "a" {
        t.Fatalf("string desc wrong: %v %v", s1, s2)
    }
}

func TestMerge_Desc_Float_Partitioned(t *testing.T) {
    var p Plan
    p.PartitionBy = "p"
    p.Sort = []SortKey{{Field:"v", Order:"desc"}}
    op := NewOperator(p)
    _ = op.Push(E(map[string]any{"p":"x", "v": 1.5}))
    _ = op.Push(E(map[string]any{"p":"x", "v": 2.7}))
    _ = op.Push(E(map[string]any{"p":"y", "v": 10.0}))
    _ = op.Push(E(map[string]any{"p":"y", "v": 7.0}))
    // collect until empty; verify per-partition desc
    var xs []float64
    var ys []float64
    for {
        e, ok := op.Pop(); if !ok { break }
        m := e.Payload.(map[string]any)
        if m["p"].(string) == "x" { xs = append(xs, m["v"].(float64)) } else { ys = append(ys, m["v"].(float64)) }
    }
    if !(len(xs) == 2 && xs[0] >= xs[1]) { t.Fatalf("x desc wrong: %v", xs) }
    if !(len(ys) == 2 && ys[0] >= ys[1]) { t.Fatalf("y desc wrong: %v", ys) }
}
