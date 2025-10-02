package kvstore

// Del deletes a key from the default store.
func Del(key string) bool { return defaultStore.Del(key) }

