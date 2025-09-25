package schemas

import "errors"

// ObjIndexV1 is a non-debug build artifact index written under
// build/obj/<package>/index.json to list per-unit assembly outputs.
// It mirrors asm.v1 index shape but is specific to non-debug obj files.
type ObjIndexV1 struct {
    Schema    string      `json:"schema"`
    Timestamp string      `json:"timestamp"`
    Package   string      `json:"package"`
    Files     []ObjFile   `json:"files"`
}

type ObjFile struct {
    Unit   string `json:"unit"`
    Path   string `json:"path"`
    Size   int64  `json:"size"`
    Sha256 string `json:"sha256"`
}

func (o *ObjIndexV1) Validate() error {
    if o == nil { return errors.New("nil obj index") }
    if o.Schema == "" { o.Schema = "objindex.v1" }
    if o.Schema != "objindex.v1" { return errors.New("invalid schema") }
    if o.Package == "" { return errors.New("package required") }
    return nil
}

