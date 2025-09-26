package schemas

import "errors"

func (b *BuildPlanV1) Validate() error {
    if b == nil {
        return errors.New("nil build plan")
    }
    if b.Schema == "" {
        b.Schema = "buildplan.v1"
    }
    if b.Schema != "buildplan.v1" {
        return errors.New("invalid schema")
    }
    if b.Workspace == "" {
        return errors.New("workspace required")
    }
    return nil
}

