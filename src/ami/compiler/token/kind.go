package token

// Kind enumerates lexical token kinds produced by the scanner.
// Use String() for stable debug names.
type Kind int

// Token kinds. Grouped by category; values are stable only within tests.
const (
    // Special
    Unknown Kind = iota
    EOF

    // Identifiers and literals
    Ident
    Number
    String
    Symbol // generic symbol bucket for single-character tokens in early scaffolds

    // Operators (subset for precedence scaffolding)
    Assign // =
    Plus   // +
    Minus  // -
    Star   // *
    Slash  // /
    Percent// %
    Bang   // !

    Eq  // ==
    Ne  // !=
    Lt  // <
    Gt  // >
    Le  // <=
    Ge  // >=

    And // &&
    Or  // ||

    Arrow // ->

    // Keywords (reserved)
    KwPackage
    KwImport
    KwFunc
    KwReturn
    KwVar
    KwDefer
    KwPipeline
    KwIngress
    KwEgress
    KwError

    // Punctuation (symbols with distinct kinds)
    LParenSym   // (
    RParenSym   // )
    LBraceSym   // {
    RBraceSym   // }
    LBracketSym // [
    RBracketSym // ]
    CommaSym    // ,
    SemiSym     // ;
    DotSym      // .
    ColonSym    // :

    // Comments
    LineComment  // //...
    BlockComment // /*...*/
)
