package rand

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers in the range [0,n).
func (p *PRNG) Perm(n int) []int { return p.r.Perm(n) }

