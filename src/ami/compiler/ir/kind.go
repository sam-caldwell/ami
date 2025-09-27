package ir

// Kind identifies the IR operation.
type Kind int

const (
    OpVar Kind = iota
    OpAssign
    OpReturn
    OpDefer
    OpExpr
)

func (k Kind) String() string {
    switch k {
    case OpVar:
        return "VAR"
    case OpAssign:
        return "ASSIGN"
    case OpReturn:
        return "RETURN"
    case OpDefer:
        return "DEFER"
    case OpExpr:
        return "EXPR"
    default:
        return "?"
    }
}

