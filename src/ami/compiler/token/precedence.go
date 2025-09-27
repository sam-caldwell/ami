package token

// Precedence levels (higher number binds tighter). Non-operator kinds return 0.
const (
    precNone = 0
    precOr   = 1
    precAnd  = 2
    precEq   = 3
    precRel  = 4
    precAdd  = 5
    precMul  = 6
)

// Precedence returns the binding power for the given operator token kind.
func Precedence(k Kind) int {
    switch k {
    case Or:
        return precOr
    case And:
        return precAnd
    case Eq, Ne:
        return precEq
    case Lt, Gt, Le, Ge:
        return precRel
    case Plus, Minus:
        return precAdd
    case Star, Slash, Percent:
        return precMul
    default:
        return precNone
    }
}

