package kvstore

// default store for convenience and testing
var defaultStore = New()

// Default returns the process-global default store instance.
func Default() *Store { return defaultStore }

// ResetDefault resets the default store for tests.
func ResetDefault() { defaultStore = New() }

