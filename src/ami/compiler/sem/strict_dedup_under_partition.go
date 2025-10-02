package sem

import (
    "os"
    "strings"
)

// StrictDedupUnderPartition controls whether certain Dedup vs PartitionBy misconfigurations
// are treated as errors rather than warnings. It can be toggled via the environment variable
// AMI_STRICT_DEDUP_PARTITION (values: "1", "true"). Tests in package sem may also set the
// variable directly.
var StrictDedupUnderPartition bool

func init() {
    v := strings.ToLower(strings.TrimSpace(os.Getenv("AMI_STRICT_DEDUP_PARTITION")))
    if v == "1" || v == "true" { StrictDedupUnderPartition = true }
}

