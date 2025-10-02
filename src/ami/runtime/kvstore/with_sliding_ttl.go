package kvstore

// WithSlidingTTL refreshes expiration to now+ttl on each successful Get.
func WithSlidingTTL() PutOption { return func(o *putOptions) { o.sliding = true } }

