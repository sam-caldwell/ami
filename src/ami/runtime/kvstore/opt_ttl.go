package kvstore

import "time"

// TTL sets an absolute expiration duration (no sliding refresh).
func TTL(d time.Duration) PutOption { return func(c *putConfig) { c.ttl = d; c.sliding = false } }

