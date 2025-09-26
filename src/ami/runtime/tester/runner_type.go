package tester

import (
    kv "github.com/sam-caldwell/ami/src/ami/runtime/kvstore"
    mem "github.com/sam-caldwell/ami/src/ami/runtime/memory"
    "sync/atomic"
)

// Runner provides a deterministic runtime executor for AMI pipelines.
// For Phase 2 scaffold, execution is simulated.
type Runner struct {
    // Mem provides per-VM memory accounting across test case execution.
    // Domains: Event heap, Node-state heap, Ephemeral stack.
    Mem *mem.Manager
    // AutoEmitKV, when true, emits kvstore.metrics at the end of Execute.
    AutoEmitKV bool
    kvMetrics  uint64 // count of kv metrics emissions
}

// New constructs a Runner with a fresh memory manager.
func New() *Runner { return &Runner{Mem: mem.NewManager()} }

// EnableAutoEmitKV toggles automatic kvstore metrics emission after run.
func (r *Runner) EnableAutoEmitKV(enable bool) { r.AutoEmitKV = enable }

// KVMetricsEmitted reports whether any kv metrics emissions occurred.
func (r *Runner) KVMetricsEmitted() bool { return atomic.LoadUint64(&r.kvMetrics) > 0 }

// KVMetricsCount returns the number of kv metrics emissions observed by this runner.
func (r *Runner) KVMetricsCount() uint64 { return atomic.LoadUint64(&r.kvMetrics) }

// snapshotKV enumerates stores; kept here to avoid cycles.
func snapshotKV() []kv.StoreInfo { return kv.Default().Snapshot() }

