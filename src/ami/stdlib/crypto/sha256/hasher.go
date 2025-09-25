package sha256

import (
    stdsha256 "crypto/sha256"
    "hash"
)

// Hasher provides streaming SHA-256 hashing.
type Hasher struct{ h hash.Hash }

// New returns a new Hasher instance.
func New() *Hasher { return &Hasher{h: stdsha256.New()} }

// Write adds more data to the running hash.
func (s *Hasher) Write(p []byte) (int, error) { return s.h.Write(p) }

// Sum returns the 32-byte checksum of the data written so far.
func (s *Hasher) Sum() [32]byte {
    sum := s.h.Sum(nil)
    var out [32]byte
    copy(out[:], sum)
    return out
}

// Reset resets the Hasher to its initial state.
func (s *Hasher) Reset() { s.h.Reset() }

