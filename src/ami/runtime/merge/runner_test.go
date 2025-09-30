package merge

import (
    "context"
    "testing"
    "time"
    ev "github.com/sam-caldwell/ami/src/schemas/events"
    "github.com/sam-caldwell/ami/src/ami/runtime/scheduler"
)

func TestMergeTask_WithScheduler_FairPolicy(t *testing.T) {
    in := make(chan ev.Event, 8)
    out := make(chan ev.Event, 8)
    var p Plan
    p.Sort = []SortKey{{Field:"k", Order:"asc"}}
    p.PartitionBy = "p"
    task := MergeTask("collect", p, in, out)
    pool, err := scheduler.New(scheduler.Config{Workers: 1, Policy: scheduler.FAIR})
    if err != nil { t.Fatalf("pool: %v", err) }
    defer pool.Stop()
    if err := pool.Submit(task); err != nil { t.Fatalf("submit: %v", err) }
    // enqueue two partitions
    in <- E(map[string]any{"p":"a", "k":2})
    in <- E(map[string]any{"p":"b", "k":1})
    in <- E(map[string]any{"p":"a", "k":1})
    // allow processing
    time.Sleep(20 * time.Millisecond)
    // collect outputs; ensure both partitions appear and 'a' is sorted locally
    var gotA []int
    var seenB bool
    for i := 0; i < 3; i++ {
        select{
        case e := <-out:
            mp := e.Payload.(map[string]any)
            if mp["p"].(string) == "a" { gotA = append(gotA, mp["k"].(int)) } else if mp["p"].(string) == "b" { seenB = true }
        default:
        }
    }
    if !seenB || len(gotA) < 2 { t.Fatalf("expected both partitions and two for 'a': a=%v seenB=%v", gotA, seenB) }
    if !(gotA[0] <= gotA[1]) { t.Fatalf("'a' not sorted: %v", gotA) }
    _ = context.Background()
}

