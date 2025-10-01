package os

import (
    goos "os"
    "os/exec"
    "runtime"
    "strconv"
    "strings"
)

// SystemStats describes basic hardware/OS attributes.
type SystemStats struct {
    OS                string
    Arch              string
    NumCPU            int
    TotalMemoryBytes  uint64
}

// SystemStats returns basic system information. Total memory is best-effort and
// may be zero if not available on this platform.
func GetSystemStats() SystemStats {
    st := SystemStats{OS: runtime.GOOS, Arch: runtime.GOARCH, NumCPU: runtime.NumCPU()}
    // Try to get total memory by platform
    switch runtime.GOOS {
    case "linux":
        if b, err := goos.ReadFile("/proc/meminfo"); err == nil {
            // Look for "MemTotal:       16341656 kB"
            lines := strings.Split(string(b), "\n")
            for _, ln := range lines {
                if strings.HasPrefix(ln, "MemTotal:") {
                    f := strings.Fields(ln)
                    if len(f) >= 2 {
                        if kb, err := strconv.ParseUint(f[1], 10, 64); err == nil {
                            st.TotalMemoryBytes = kb * 1024
                        }
                    }
                    break
                }
            }
        }
    case "darwin":
        // sysctl -n hw.memsize
        if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
            s := strings.TrimSpace(string(out))
            if n, err := strconv.ParseUint(s, 10, 64); err == nil { st.TotalMemoryBytes = n }
        }
    // other platforms: leave zero
    }
    return st
}
