package kvstore

import "time"

// PutOption configures behavior for a single Put operation.
type PutOption func(*putConfig)

type putConfig struct {
    ttl       time.Duration
    sliding   bool
    maxReads  int // <=0 means unlimited
}

// TTL sets an absolute expiration duration (no sliding refresh).
func TTL(d time.Duration) PutOption { return func(c *putConfig) { c.ttl = d; c.sliding = false } }

// SlidingTTL sets a sliding expiration duration (deadline refreshes on access).
func SlidingTTL(d time.Duration) PutOption { return func(c *putConfig) { c.ttl = d; c.sliding = true } }

// MaxReads limits the number of Get() calls that can retrieve the value.
// After N reads, the entry is deleted. Use 1 for one-time reads.
func MaxReads(n int) PutOption { return func(c *putConfig) { c.maxReads = n } }

// Options configures a Store instance.
type Options struct {
    // CapacityBytes sets an approximate memory cap for all entries in this store.
    // When exceeded, the store evicts least-recently-used entries until under the cap.
    // A value <= 0 disables capacity limiting.
    CapacityBytes int64

    // SweepInterval configures how often a background janitor scans for TTL expirations.
    // A value <= 0 disables the janitor; expirations will be enforced lazily on access.
    SweepInterval time.Duration
}

