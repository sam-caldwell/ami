package driver

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

func opName(k token.Kind) string {
    switch k {
    case token.Plus: return "add"
    case token.Minus: return "sub"
    case token.Star: return "mul"
    case token.Slash: return "div"
    case token.Percent: return "mod"
    case token.And: return "and"
    case token.Or:  return "or"
    case token.BitXor: return "xor"
    case token.BitOr: return "bor"
    case token.Shl: return "shl"
    case token.Shr: return "shr"
    case token.BitAnd: return "band"
    case token.Eq: return "eq"
    case token.Ne: return "ne"
    case token.Lt: return "lt"
    case token.Le: return "le"
    case token.Gt: return "gt"
    case token.Ge: return "ge"
    default:
        return k.String()
    }
}

