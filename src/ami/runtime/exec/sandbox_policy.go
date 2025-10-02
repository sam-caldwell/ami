package exec

// SandboxPolicy controls which capabilities are allowed for ingress/source stages.
// It is a cooperative simulation hook for tests and examples; not an OS sandbox.
type SandboxPolicy struct {
    AllowFS     bool
    AllowNet    bool
    AllowDevice bool
}

func (p SandboxPolicy) allow(cap string) bool {
    switch cap {
    case "fs":
        return p.AllowFS
    case "net":
        return p.AllowNet
    case "device":
        return p.AllowDevice
    default:
        return true
    }
}

