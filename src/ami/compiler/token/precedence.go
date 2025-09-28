package token

// Precedence levels (higher number binds tighter). Non-operator kinds return 0.
const (
    precNone = 0
    precOr   = 1
    precAnd  = 2
    // Bitwise levels (tighter than logical, looser than arithmetic/rel/eq per simplified model)
    precBitOr  = 3
    precBitXor = 4
    precBitAnd = 5
    precEq   = 6
    precRel  = 7
    // Shift placed near arithmetic; matches simplified left-associative grammar
    precAdd  = 8
    precMul  = 9
    precShift = 8
)

// Precedence returns the binding power for the given operator token kind.
func Precedence(k Kind) int {
    switch k {
    case Or:
        return precOr
    case And:
        return precAnd
    case BitOr:
        return precBitOr
    case BitXor:
        return precBitXor
    case BitAnd:
        return precBitAnd
    case Eq, Ne:
        return precEq
    case Lt, Gt, Le, Ge:
        return precRel
    case Shl, Shr:
        return precShift
    case Plus, Minus:
        return precAdd
    case Star, Slash, Percent:
        return precMul
    default:
        return precNone
    }
}
