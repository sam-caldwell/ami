package kvstore

import lg "github.com/sam-caldwell/ami/src/internal/logger"

// EmitMetrics logs a diag.v1 record for the store's current metrics.
// Optional pipeline/node names can be included for attribution.
func (s *Store) EmitMetrics(pipeline, node string) {
	m := s.Metrics()
	data := map[string]interface{}{
		"hits":        m.Hits,
		"misses":      m.Misses,
		"expirations": m.Expirations,
		"evictions":   m.Evictions,
		"entries":     m.Entries,
		"bytesUsed":   m.BytesUsed,
	}
	if pipeline != "" {
		data["pipeline"] = pipeline
	}
	if node != "" {
		data["node"] = node
	}
	lg.Info("kvstore.metrics", data)
}
