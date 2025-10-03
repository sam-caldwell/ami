package testutil

import (
    "math"
    "os"
    "strconv"
    "time"
)

// Timeout scales a base duration by AMI_TEST_TIMEOUT_SCALE when set.
// Examples:
//   AMI_TEST_TIMEOUT_SCALE=2     -> doubles timeouts
//   AMI_TEST_TIMEOUT_SCALE=0.5   -> halves timeouts
// Invalid values are ignored and the base duration is returned.
func Timeout(base time.Duration) time.Duration {
    s := os.Getenv("AMI_TEST_TIMEOUT_SCALE")
    if s == "" { return base }
    f, err := strconv.ParseFloat(s, 64)
    if err != nil { return base }
    if f <= 0 || math.IsNaN(f) || math.IsInf(f, 0) { return base }
    // Guard against overflow on extreme scales
    scaled := float64(base) * f
    if scaled > float64(time.Duration(math.MaxInt64)) { return time.Duration(math.MaxInt64) }
    return time.Duration(scaled)
}

