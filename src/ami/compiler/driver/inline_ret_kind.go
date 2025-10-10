package driver

type retKind int

const (
    retNone retKind = iota
    retEV
    retLit
    retBinOp
    retCmp
    retUnary
)
