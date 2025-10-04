package os

// SystemStats describes basic hardware/OS attributes.
type SystemStats struct {
    OS                string
    Arch              string
    NumCPU            int
    TotalMemoryBytes  uint64
}

