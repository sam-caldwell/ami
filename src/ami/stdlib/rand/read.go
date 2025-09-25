package rand

// Read generates len(p) random bytes and writes them into p. It always returns len(p) and a nil error.
func (p *PRNG) Read(b []byte) (int, error) { return p.r.Read(b) }
