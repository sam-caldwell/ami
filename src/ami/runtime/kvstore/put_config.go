package kvstore

import "time"

type putConfig struct {
    ttl      time.Duration
    sliding  bool
    maxReads int // <=0 means unlimited
}

