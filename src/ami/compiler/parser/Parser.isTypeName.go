package parser

import "github.com/sam-caldwell/ami/src/ami/compiler/token"

// isTypeName returns true if the current token kind is a valid type name token
// (an identifier or a recognized primitive type keyword).
func (p *Parser) isTypeName(k token.Kind) bool {
	switch k {
	case token.Ident,
		token.KwBool,
		token.KwByte, token.KwInt, token.KwInt8, token.KwInt16, token.KwInt32, token.KwInt64, token.KwInt128,
		token.KwUint, token.KwUint8, token.KwUint16, token.KwUint32, token.KwUint64, token.KwUint128,
		token.KwFloat32, token.KwFloat64,
		token.KwStringTy, token.KwRune,
		token.KwSlice,
		token.KwSet, token.KwMap,
		token.KwStruct,
		token.KwEvent, token.KwError:
		return true
	default:
		return false
	}
}
