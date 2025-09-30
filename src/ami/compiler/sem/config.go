package sem

// SetStrictDedupUnderPartition sets the package-level toggle for elevating
// Dedup under PartitionBy warnings to errors.
func SetStrictDedupUnderPartition(v bool) { StrictDedupUnderPartition = v }

