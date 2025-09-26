package schemas

import "errors"

func (e *EdgesV1) Validate() error {
    if e == nil { return errors.New("nil edges") }
    if e.Schema == "" { e.Schema = "edges.v1" }
    if e.Schema != "edges.v1" { return errors.New("invalid schema") }
    return nil
}

