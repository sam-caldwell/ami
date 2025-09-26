package token

// Lexical rune and string constants recognized by the scanner.
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
