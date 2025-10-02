package kvstore

// WithMaxReads sets delete-on-read after N successful reads.
func WithMaxReads(n int) PutOption { return func(o *putOptions) { if n > 0 { o.maxReads = n } } }

