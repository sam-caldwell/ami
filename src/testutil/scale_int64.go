package testutil

import (
    "math"
    "os"
    "strconv"
)

// ScaleInt64 applies the same scaling as ScaleInt but for int64 values.
func ScaleInt64(n int64) int64 {
    if n <= 0 { return n }
    s := os.Getenv("AMI_TEST_TIMEOUT_SCALE")
    if s == "" { return n }
    f, err := strconv.ParseFloat(s, 64)
    if err != nil || f <= 0 || math.IsNaN(f) || math.IsInf(f, 0) { return n }
    scaled := int64(math.Ceil(float64(n) * f))
    if scaled < n { // overflow guard
        return math.MaxInt64
    }
    return scaled
}

