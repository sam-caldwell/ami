package llvm

import (
    "os"
    "strings"
)

// gpuAllowedBackends inspects the AMI_GPU_BACKENDS env var and returns booleans for
// allowing metal, cuda, opencl in the generated runtime IR. Empty or missing means all allowed.
func gpuAllowedBackends() (metal, cuda, opencl bool) {
    v := os.Getenv("AMI_GPU_BACKENDS")
    if v == "" {
        return true, true, true
    }
    var m, c, o bool
    parts := strings.Split(v, ",")
    for _, p := range parts {
        p = strings.TrimSpace(strings.ToLower(p))
        switch p {
        case "metal":
            m = true
        case "cuda":
            c = true
        case "opencl", "ocl":
            o = true
        }
    }
    return m, c, o
}

