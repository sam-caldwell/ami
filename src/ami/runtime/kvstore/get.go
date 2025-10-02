package kvstore

// Get retrieves a value from the default store.
func Get(key string) (any, bool) { return defaultStore.Get(key) }

