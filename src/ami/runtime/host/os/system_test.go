package os

import (
    "runtime"
    "testing"
)

func TestSystemStats_BasicFields(t *testing.T) {
    st := GetSystemStats()
    if st.OS != runtime.GOOS { t.Fatalf("OS mismatch: %s vs %s", st.OS, runtime.GOOS) }
    if st.Arch != runtime.GOARCH { t.Fatalf("Arch mismatch: %s vs %s", st.Arch, runtime.GOARCH) }
    if st.NumCPU <= 0 { t.Fatalf("NumCPU should be >0; got %d", st.NumCPU) }
    // TotalMemoryBytes is best-effort; allow zero on unsupported platforms
    if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
        if st.TotalMemoryBytes == 0 { t.Fatalf("expected non-zero total memory on %s", runtime.GOOS) }
    }
}
