package kvstore

// NewRegistry creates a new Registry with Store options used for new instances.
func NewRegistry(opts Options) *Registry {
    return &Registry{opts: opts, table: make(map[string]*Store)}
}

