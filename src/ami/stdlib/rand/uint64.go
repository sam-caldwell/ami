package rand

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (p *PRNG) Uint64() uint64 { return p.r.Uint64() }

