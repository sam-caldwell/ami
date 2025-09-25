package token

type Kind int

const (
	// Special
	EOF Kind = iota
	ILLEGAL

	// Identifiers and literals
	IDENT
	NUMBER
	STRING

	// Keywords (subset; expand as language grows)
	KW_PACKAGE
	KW_IMPORT
	KW_FUNC
	KW_ENUM
	KW_STRUCT
	KW_MAP
	KW_SET
	KW_SLICE
	KW_INGRESS
	KW_TRANSFORM
	KW_FANOUT
	KW_COLLECT
	KW_EGRESS
	KW_ERROR
	KW_PIPELINE
	KW_STATE
	KW_TRUE
	KW_FALSE
	KW_NIL
	KW_MUT
	KW_DEFER
	KW_RETURN
	KW_VAR

	// Operators and delimiters
	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
	LBRACK // [
	RBRACK // ]
	COMMA  // ,
	SEMI   // ;
	COLON  // :
	DOT    // .

	ASSIGN  // =
	ARROW   // ->
	PIPE    // |
	EQ      // ==
	NEQ     // !=
	LT      // <
	LTE     // <=
	GT      // >
	GTE     // >=
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %
	AMP     // &

	// Directives
	PRAGMA // lexer-level: lines beginning with #pragma
)
