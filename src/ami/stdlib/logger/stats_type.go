package logger

// Stats returns a snapshot of pipeline counters.
type Stats struct{ Enqueued, Written, Dropped, Batches, Flushes int64 }

