package schemas

import "errors"

type ASMIndexV1 struct {
    Schema   string     `json:"schema"`
    Timestamp string    `json:"timestamp"`
    Package  string     `json:"package"`
    Files    []ASMFile  `json:"files"`
}

type ASMFile struct {
    Unit  string `json:"unit"`
    Path  string `json:"path"`
    Size  int64  `json:"size"`
    Sha256 string `json:"sha256"`
}

func (a *ASMIndexV1) Validate() error {
    if a == nil { return errors.New("nil asm index") }
    if a.Schema == "" { a.Schema = "asm.v1" }
    if a.Schema != "asm.v1" { return errors.New("invalid schema") }
    return nil
}

