package ir

// Kind identifies the IR operation.
type Kind int

const (
    OpVar Kind = iota
    OpAssign
    OpReturn
    OpDefer
    OpExpr
    OpLoop
    OpGoto
    OpSetPC
    OpDispatch
    OpPushFrame
    OpPopFrame
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
    case OpLoop:
        return "LOOP"
    case OpGoto:
        return "GOTO"
    case OpSetPC:
        return "SET_PC"
    case OpDispatch:
        return "DISPATCH"
    case OpPushFrame:
        return "PUSH_FRAME"
    case OpPopFrame:
        return "POP_FRAME"
    default:
        return "?"
    }
}
