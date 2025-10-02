package codegen

// ObjUnit describes a single object unit entry.
type ObjUnit struct {
    Unit   string `json:"unit"`
    Path   string `json:"path"`
    Size   int64  `json:"size"`
    Sha256 string `json:"sha256"`
}

