package testutil

import (
    "math"
    "os"
    "strconv"
)

// ScaleInt scales an integer count/backoff by AMI_TEST_TIMEOUT_SCALE.
// Rounds up for positive values so that small counts do not round to zero.
// Non-positive values are returned unchanged. Invalid scales are ignored.
func ScaleInt(n int) int {
    if n <= 0 { return n }
    s := os.Getenv("AMI_TEST_TIMEOUT_SCALE")
    if s == "" { return n }
    f, err := strconv.ParseFloat(s, 64)
    if err != nil || f <= 0 || math.IsNaN(f) || math.IsInf(f, 0) { return n }
    scaled := int(math.Ceil(float64(n) * f))
    if scaled < n { // overflow guard
        return int(math.MaxInt)
    }
    return scaled
}

