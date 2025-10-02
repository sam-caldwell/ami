package kvstore

import "time"

type putOptions struct {
    ttl      time.Duration
    maxReads int
    sliding  bool
}

