package enum

// Values returns all enum ordinals in canonical order.
func Values(d Descriptor) []int {
    out := make([]int, len(d.Names))
    for i := range out { out[i] = i }
    return out
}

