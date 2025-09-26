package driver

import sch "github.com/sam-caldwell/ami/src/schemas"

// Result holds compiler outputs for scaffolding and debug artifacts.
type Result struct {
    AST       []sch.ASTV1
    IR        []sch.IRV1
    Pipelines []sch.PipelinesV1
    EventMeta []sch.EventMetaV1
    ASM       []ASMUnit // assembly text per unit
}

