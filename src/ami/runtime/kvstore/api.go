package kvstore

// Put stores a value in the default store.
func Put(key string, val any, opts ...PutOption) { defaultStore.Put(key, val, opts...) }

// Get retrieves a value from the default store.
func Get(key string) (any, bool) { return defaultStore.Get(key) }

// Del deletes a key from the default store.
func Del(key string) bool { return defaultStore.Del(key) }

// Has returns true if key exists (not expired) in the default store.
func Has(key string) bool { return defaultStore.Has(key) }

// Keys returns all keys in the default store.
func Keys() []string { return defaultStore.Keys() }

// Stats returns default store metrics snapshot.
func Stats() Metrics { return defaultStore.Metrics() }

// SetCapacity sets default store capacity (0 disables eviction).
func SetCapacity(n int) { defaultStore.SetCapacity(n) }
