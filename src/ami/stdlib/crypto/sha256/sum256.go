package sha256

import stdsha256 "crypto/sha256"

// Sum256 returns the SHA-256 checksum of the data.
func Sum256(data []byte) [32]byte { return stdsha256.Sum256(data) }

