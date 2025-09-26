package manifest

// Manifest is the top-level workspace manifest schema.
type Manifest struct {
    Schema    string     `json:"schema"`
    Project   Project    `json:"project"`
    Packages  []Package  `json:"packages"`
    Artifacts []Artifact `json:"artifacts"`
    Toolchain Toolchain  `json:"toolchain"`
    CreatedAt string     `json:"createdAt"`
}

// Project captures the workspace name and version.
type Project struct {
    Name    string `json:"name"`
    Version string `json:"version"`
}

// Package references a dependency with version and digest.
type Package struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Digest  string `json:"digestSHA256"`
    Source  string `json:"source"`
}

// Artifact describes a build output with path and digest.
type Artifact struct {
    Path   string `json:"path"`
    Kind   string `json:"kind"`
    Size   int64  `json:"size"`
    Sha256 string `json:"sha256"`
}

// Toolchain captures tool versions used for the build.
type Toolchain struct {
    AmiVersion string `json:"amiVersion"`
    GoVersion  string `json:"goVersion"`
}

