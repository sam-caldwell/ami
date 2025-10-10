package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// isAttrBareValueKeyword reports whether a token kind is a bare keyword form
// that should be accepted as a value in key=value attribute arguments without
// requiring full expression parsing. This intentionally excludes plain
// identifiers to avoid prematurely consuming qualified calls like edge.FIFO(...).
func (p *Parser) isAttrBareValueKeyword(k token.Kind) bool {
    switch k {
    case token.KwBool,
        token.KwByte, token.KwInt, token.KwInt8, token.KwInt16, token.KwInt32, token.KwInt64, token.KwInt128,
        token.KwUint, token.KwUint8, token.KwUint16, token.KwUint32, token.KwUint64, token.KwUint128,
        token.KwFloat32, token.KwFloat64,
        token.KwStringTy, token.KwRune,
        token.KwTrue, token.KwFalse,
        token.KwError, token.KwErrorEvent, token.KwErrorPipeline:
        return true
    default:
        return false
    }
}

