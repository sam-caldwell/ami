package rand

import (
    stdrand "math/rand"
)

// PRNG is a deterministic pseudo-random generator with an explicit seed.
type PRNG struct { r *stdrand.Rand }

// New returns a new deterministic PRNG seeded with seed.
func New(seed int64) *PRNG { return &PRNG{ r: stdrand.New(stdrand.NewSource(seed)) } }

