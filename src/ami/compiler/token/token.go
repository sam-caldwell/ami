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
    LPAREN    // (
    RPAREN    // )
    LBRACE    // {
    RBRACE    // }
    LBRACK    // [
    RBRACK    // ]
    COMMA     // ,
    SEMI      // ;
    COLON     // :
    DOT       // .

    ASSIGN    // =
    ARROW     // ->
    PIPE      // |
    EQ        // ==
    NEQ       // !=
    LT        // <
    LTE       // <=
    GT        // >
    GTE       // >=
    PLUS      // +
    MINUS     // -
    STAR      // *
    SLASH     // /
    PERCENT   // %
    AMP       // &

    // Directives
    PRAGMA // lexer-level: lines beginning with #pragma
)

type Token struct {
    Kind   Kind
    Lexeme string
    Line   int
    Column int
    Offset int
}

// Keywords maps the lower‑case source lexeme to its keyword token kind.
var Keywords = map[string]Kind{
    "package":   KW_PACKAGE,
    "import":    KW_IMPORT,
    "func":      KW_FUNC,
    "enum":      KW_ENUM,
    "struct":    KW_STRUCT,
    "map":       KW_MAP,
    "set":       KW_SET,
    "slice":     KW_SLICE,
    "ingress":   KW_INGRESS,
    "transform": KW_TRANSFORM,
    "fanout":    KW_FANOUT,
    "collect":   KW_COLLECT,
    "egress":    KW_EGRESS,
    "error":     KW_ERROR,
    "pipeline":  KW_PIPELINE,
    "state":     KW_STATE,
    "true":      KW_TRUE,
    "false":     KW_FALSE,
    "nil":       KW_NIL,
    "mut":       KW_MUT,
    "defer":     KW_DEFER,
    "return":    KW_RETURN,
    "var":       KW_VAR,
}
