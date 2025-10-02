package kvstore

// Has returns true if key exists (not expired) in the default store.
func Has(key string) bool { return defaultStore.Has(key) }

