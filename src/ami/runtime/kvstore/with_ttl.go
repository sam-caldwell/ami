package kvstore

import "time"

// WithTTL sets absolute time-to-live duration for the key.
func WithTTL(d time.Duration) PutOption { return func(o *putOptions) { o.ttl = d } }

