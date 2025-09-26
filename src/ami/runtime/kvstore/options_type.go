package kvstore

import "time"

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

