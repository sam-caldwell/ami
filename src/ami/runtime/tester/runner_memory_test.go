package tester

import (
	mem "github.com/sam-caldwell/ami/src/ami/runtime/memory"
	"testing"
)

// The Runner wires a per-VM memory manager; allocations for each case are
// released deterministically after Execute returns (RAII-like behavior).
func TestRunner_MemoryManager_StatsZeroAfterExecute(t *testing.T) {
	r := New()
	cases := []Case{
		{Name: "A", Pipeline: "P", InputJSON: `{"x":1}`, ExpectJSON: `{"x":1}`},
		{Name: "B", Pipeline: "P", InputJSON: `{"sleep_ms":1}`, ExpectJSON: `{}`},
	}
	_, err := r.Execute("P", cases)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	st := r.Mem.Stats()
	if st[mem.Event] != 0 || st[mem.State] != 0 || st[mem.Ephemeral] != 0 {
		t.Fatalf("memory not fully released: %+v", st)
	}
}
