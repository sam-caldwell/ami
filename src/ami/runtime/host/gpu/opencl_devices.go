package gpu

// OpenCLDevices enumerates devices for a given platform (env-forced dummy).
func OpenCLDevices(p Platform) []Device {
    if !OpenCLAvailable() { return nil }
    if p.Name == "" && p.Vendor == "" && p.Version == "" { return nil }
    return []Device{{Backend: "opencl", ID: 0, Name: "opencl-ci-0"}}
}

