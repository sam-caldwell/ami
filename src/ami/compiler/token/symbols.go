package token

// Lexical rune and string constants recognized by the scanner.
const (
	LexAmpersand   = '&'
	LexAsterisk    = '*'
	LexBkSlash     = '\\' // skip escape and next
	LexColon       = ':'
	LexComma       = ','
	LexCr          = '\r'
	LexDblQuote    = '"'
	LexEQ          = '='
	LexExclamation = '!'
	LexGt          = '>'
	LexHyphen      = '-'
	LexLBrace      = '{'
	LexLBracket    = '['
	LexLf          = '\n'
	LexLParen      = '('
	LexLt          = '<'
	LexPercent     = '%'
	LexPeriod      = '.'
	LexPipe        = '|'
	LexPlus        = '+'
	LexRBrace      = '}'
	LexRBracket    = ']'
	LexRParen      = ')'
	LexSemicolon   = ';'
	LexSlash       = '/'
	LexSpace       = ' '
	LexTab         = '\t'

	LexUnderscore = '_'
)

const (
	LexArrowRight = "->"
	LexBoolEQ     = "=="
	LexBoolNE     = "!="
	LexBoolGT     = ">"
	LexBoolGE     = ">="
	LexBoolLT     = "<"
	LexBoolLE     = "<="

	LexPragma = "#pragma"
)
