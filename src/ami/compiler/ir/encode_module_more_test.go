package ir

import "testing"

func TestEncodeModule_MoreBranches(t *testing.T) {
    m := Module{
        Package:     "app",
        Concurrency: 4,
        Backpressure: "dropNewest",
        TelemetryEnabled: true,
        Schedule:    "fifo",
        Capabilities: []string{"io","net"},
        TrustLevel:  "trusted",
        ExecContext: &ExecContext{Env: map[string]string{"K":"V"}},
        EventMeta:   &EventMeta{Schema: "ev.v1", Fields: []string{"ts","id"}},
        Directives:  []Directive{{Domain:"lint", Key:"disable", Value:"X", Args: []string{"a","b"}, Params: map[string]string{"k":"v"}}},
        Pipelines: []Pipeline{{Name:"P", Collect: []CollectSpec{{Step:"Collect", Merge:&MergePlan{Stable:true, Sort: []SortKey{{Field:"k", Order:"asc"}}, Key:"id", PartitionBy:"p", Buffer: BufferPlan{Capacity:10, Policy:"dropNewest"}, Window:2, TimeoutMs:10, DedupField:"id", Watermark:&Watermark{Field:"ts", LatenessMs: 5}, LatePolicy:"drop"}}}}},
        Functions: []Function{{Name:"f", Params: []Value{{ID:"a", Type:"int"}}, Results: []Value{{ID:"r", Type:"int"}}, Decorators: []Decorator{{Name:"dec", Args: []string{"x"}}}, Blocks: []Block{{Name:"entry", Instr: []Instruction{Assign{DestID:"x", Src: Value{ID:"a", Type:"int"}}}}}}},
    }
    b, err := EncodeModule(m)
    if err != nil || len(b) == 0 { t.Fatalf("encode: %v", err) }
}
