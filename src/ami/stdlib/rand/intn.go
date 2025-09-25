package rand

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (p *PRNG) Intn(n int) int { return p.r.Intn(n) }
