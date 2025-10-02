package kvstore

// SetCapacity sets default store capacity (0 disables eviction).
func SetCapacity(n int) { defaultStore.SetCapacity(n) }

