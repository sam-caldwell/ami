package token

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
