package token

// Symbol constants capture common lexemes used by the language.
// Keep this declarative and in sync with SPEC 2.1.
const (
    // Delimiters
    LParen = "("
    RParen = ")"
    LBrace = "{"
    RBrace = "}"
    LBracket = "["
    RBracket = "]"

    // Punctuation
    Comma  = ","
    Semi   = ";"
    Colon  = ":"
    Dot    = "."
    Pipe   = "|"

    // Operators/Symbols
    AssignSym      = "="
    EqSym          = "=="
    NeSym          = "!="
    LtSym          = "<"
    LeSym          = "<="
    GtSym          = ">"
    GeSym          = ">="
    PlusSym        = "+"
    HyphenSym      = "-"
    AsteriskSym    = "*"
    SlashSym       = "/"
    PercentSym     = "%"
    BackSlash      = "\\"
    Dollar         = "$"
    Exclamation    = "!"
    Tick           = "`"
    Tilde          = "~"
    Question       = "?"
    At             = "@"
    Pound          = "#"
    Caret          = "^"
    DoubleQuote    = "\""
    SingleQuote    = "'"
)
