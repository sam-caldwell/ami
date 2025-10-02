package kvstore

// Stats returns default store metrics snapshot.
func Stats() Metrics { return defaultStore.Metrics() }

