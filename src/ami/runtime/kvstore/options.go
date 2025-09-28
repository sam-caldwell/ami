package kvstore

import "time"

// PutOption configures Put behavior.
type PutOption func(*putOptions)

type putOptions struct {
    ttl      time.Duration
    maxReads int
}

func applyOptions(opts []PutOption) putOptions {
    var o putOptions
    for _, fn := range opts { fn(&o) }
    return o
}

// WithTTL sets absolute time-to-live duration for the key.
func WithTTL(d time.Duration) PutOption { return func(o *putOptions) { o.ttl = d } }

// WithMaxReads sets delete-on-read after N successful reads.
func WithMaxReads(n int) PutOption { return func(o *putOptions) { if n > 0 { o.maxReads = n } } }

