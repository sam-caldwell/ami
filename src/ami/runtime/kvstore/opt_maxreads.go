package kvstore

// MaxReads limits the number of Get() calls that can retrieve the value.
// After N reads, the entry is deleted. Use 1 for one-time reads.
func MaxReads(n int) PutOption { return func(c *putConfig) { c.maxReads = n } }

