package kvstore

import (
    "encoding/json"
    "io"
    "sort"
    "time"
    
    "github.com/sam-caldwell/ami/src/schemas/diag"
)

// MetricsRecord builds a diag.v1 record for this store's metrics snapshot.
// Namespace may be empty for the default store.
func (s *Store) MetricsRecord(namespace string) diag.Record {
    m := s.Metrics()
    return diag.Record{
        Timestamp: time.Unix(0, 0).UTC(),
        Level:     diag.Info,
        Code:      "KV_METRICS",
        Message:   "kvstore metrics snapshot",
        Data: map[string]any{
            "namespace":   namespace,
            "hits":        m.Hits,
            "misses":      m.Misses,
            "expirations": m.Expirations,
            "evictions":   m.Evictions,
            "currentSize": m.CurrentSize,
        },
    }
}

// EmitRegistryMetrics writes one diag.v1 JSON object per registered namespace to w, sorted by namespace.
func EmitRegistryMetrics(w io.Writer) error {
    defaultRegistry.mu.RLock()
    keys := make([]string, 0, len(defaultRegistry.stores))
    for ns := range defaultRegistry.stores { keys = append(keys, ns) }
    defaultRegistry.mu.RUnlock()
    sort.Strings(keys)
    enc := json.NewEncoder(w)
    // emit default store metrics first (empty namespace), then namespaced stores
    if err := enc.Encode(defaultStore.MetricsRecord("")); err != nil { return err }
    for _, ns := range keys {
        s := Namespace(ns)
        if err := enc.Encode(s.MetricsRecord(ns)); err != nil { return err }
    }
    return nil
}

