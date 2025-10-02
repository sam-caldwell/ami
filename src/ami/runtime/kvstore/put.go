package kvstore

// Put stores a value in the default store.
func Put(key string, val any, opts ...PutOption) { defaultStore.Put(key, val, opts...) }

