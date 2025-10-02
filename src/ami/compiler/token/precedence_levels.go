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

