package ir

// GPUBlock describes a GPU kernel block captured from source with attributes.
type GPUBlock struct {
    Family string   // metal|cuda|opencl
    Name   string   // kernel entry name (optional)
    Source string   // raw kernel source
    N      int      // optional element count
    Grid   [3]int   // grid dims
    TPG    [3]int   // threads per group dims
    Args   string   // optional argument binding spec (e.g., "1buf1u32", "1buf")
}
