package kvstore

import "time"

// SlidingTTL sets a sliding expiration duration (deadline refreshes on access).
func SlidingTTL(d time.Duration) PutOption { return func(c *putConfig) { c.ttl = d; c.sliding = true } }

