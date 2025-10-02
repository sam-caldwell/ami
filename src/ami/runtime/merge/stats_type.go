package merge

// Stats captures operator counters for reporting.
type Stats struct{ Enqueued, Emitted, Dropped, Expired int64 }

