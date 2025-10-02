package codegen

// ObjIndex is the on-disk schema for build/obj/<package>/index.json (objindex.v1).
type ObjIndex struct {
    Schema  string    `json:"schema"`
    Package string    `json:"package"`
    Units   []ObjUnit `json:"units"`
}

