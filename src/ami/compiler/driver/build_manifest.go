package driver

// bmUnit and bmPackage moved to dedicated files

type BuildManifest struct {
    Schema   string      `json:"schema"`
    Packages []bmPackage `json:"packages"`
}
// writer moved to build_manifest_write.go to satisfy single-declaration rule
