package tester

import "testing"

func TestRunner_KVMetricsCount_MetaAndAuto(t *testing.T) {
	ResetDefaultKVForTest(t)
	r := New()
	if r.KVMetricsCount() != 0 {
		t.Fatalf("expected initial count 0; got %d", r.KVMetricsCount())
	}

	// Meta-driven emission (kv_emit=true)
	in1, err := BuildKVInput(nil, WithKV("Pc1", "Node1"), KVPut("k", "v"), KVEmit())
	if err != nil {
		t.Fatalf("compose input1: %v", err)
	}
	_, _ = r.Execute("Pc1", []Case{{Name: "Meta", Pipeline: "Pc1", InputJSON: in1}})
	c1 := r.KVMetricsCount()
	if c1 < 1 {
		t.Fatalf("expected count >=1 after meta emission; got %d", c1)
	}

	// Auto emission at end of Execute (kv_emit not set)
	r.EnableAutoEmitKV(true)
	in2, err := BuildKVInput(nil, WithKV("Pc2", "Node2"), KVPut("a", 1))
	if err != nil {
		t.Fatalf("compose input2: %v", err)
	}
	_, _ = r.Execute("Pc2", []Case{{Name: "Auto", Pipeline: "Pc2", InputJSON: in2}})
	c2 := r.KVMetricsCount()
	if c2 <= c1 {
		t.Fatalf("expected count to increase after auto emission; before=%d after=%d", c1, c2)
	}
}
