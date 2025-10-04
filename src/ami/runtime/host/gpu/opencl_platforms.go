package gpu

// OpenCLPlatforms lists OpenCL platforms. Always empty in stub.
func OpenCLPlatforms() []Platform {
    if OpenCLAvailable() {
        return []Platform{{Vendor: "CI", Name: "OpenCL-Dummy", Version: "1.2"}}
    }
    return nil
}

