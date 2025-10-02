package kvstore

// ResetRegistry resets the default registry (for tests).
func ResetRegistry() { defaultRegistry = &Registry{stores: map[string]*Store{}} }

