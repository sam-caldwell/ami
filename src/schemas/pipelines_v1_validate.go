package schemas

import "errors"

func (p *PipelinesV1) Validate() error {
    if p == nil { return errors.New("nil pipelines") }
    if p.Schema == "" { p.Schema = "pipelines.v1" }
    if p.Schema != "pipelines.v1" { return errors.New("invalid schema") }
    return nil
}

