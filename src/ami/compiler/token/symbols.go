package token

const (
	LexTab        = '\t'
	LexCr         = '\r'
	LexLf         = '\n'
	LexUnderscore = '_'
	LexSpace      = ' '
	LexPeriod     = '.'
	LexEQ         = '='
	LexDblQuote   = '"'
	LexBkSlash    = '\\' // skip escape and next

	LexBoolEQ = "=="
	LexBoolNE = "!="

	LexPragma = "#pragma"
)
