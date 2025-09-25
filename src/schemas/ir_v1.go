package schemas

import "errors"

type IRV1 struct {
    Schema    string       `json:"schema"`
    Timestamp string       `json:"timestamp"`
    Package   string       `json:"package"`
    File      string       `json:"file"`
    Functions []IRFunction `json:"functions"`
}

type IRFunction struct {
    Name   string    `json:"name"`
    Blocks []IRBlock `json:"blocks"`
}

type IRBlock struct {
    Label string    `json:"label"`
    Instrs []IRInstr `json:"instrs"`
}

type IRInstr struct {
    Op     string        `json:"op"`
    Args   []interface{} `json:"args,omitempty"`
    Result string        `json:"result,omitempty"`
}

func (i *IRV1) Validate() error {
    if i == nil { return errors.New("nil ir") }
    if i.Schema == "" { i.Schema = "ir.v1" }
    if i.Schema != "ir.v1" { return errors.New("invalid schema") }
    return nil
}

