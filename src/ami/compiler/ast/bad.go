package ast

import tok "github.com/sam-caldwell/ami/src/ami/compiler/token"

// Bad is a placeholder node used when a token could not be parsed.
type Bad struct{ Tok tok.Token }

func (Bad) isNode() {}

