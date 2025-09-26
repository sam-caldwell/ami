package schemas

// ToolchainV1 captures toolchain metadata for reproducibility.
type ToolchainV1 struct {
    AmiVersion string `json:"amiVersion"`
    GoVersion  string `json:"goVersion"`
}

