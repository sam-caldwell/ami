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
    DurationLit
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

    // Bitwise and shift operators
    BitAnd // &
    BitXor // ^
    BitOr  // |
    Shl    // <<
    Shr    // >>

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
    // Primitive/type keywords (subset used by parser as type names)
    KwBool
    KwByte
    KwInt
    KwInt8
    KwInt16
    KwInt32
    KwInt64
    KwInt128
    KwUint
    KwUint8
    KwUint16
    KwUint32
    KwUint64
    KwUint128
    KwFloat32
    KwFloat64
    KwStringTy // 'string' (distinct from String literal token)
    KwRune

    // Additional reserved words (not all used by parser yet)
    KwAppend
    KwBreak
    KwCase
    KwConst
    KwClose
    KwComplex
    KwComplex64
    KwComplex128
    KwContinue
    KwCopy
    KwDefault
    KwDelete
    KwElse
    KwEnum
    KwErrorEvent
    KwErrorPipeline
    KwEvent
    KwFallthrough
    KwFalse
    KwFloat
    KwFor
    KwGoto
    KwIf
    KwInterface
    KwLabel
    KwLatest
    KwLen
    KwMake
    KwMap
    KwSet
    KwSlice
    KwNew
    KwNil
    KwPanic
    KwRange
    KwReal
    KwRecover
    KwSelect
    KwState
    KwStruct
    KwSwitch
    KwTrue
    KwType
    // Node keywords
    KwNodeTransform
    KwNodeFanout
    KwNodeCollect
    KwNodeMutable

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
    PipeSym     // |
    BackslashSym// \
    DollarSym   // $
    TickSym     // `
    TildeSym    // ~
    QuestionSym // ?
    AtSym       // @
    PoundSym    // #
    CaretSym    // ^
    SingleQuoteSym // '
    DoubleQuoteSym // "

    // Comments
    LineComment  // //...
    BlockComment // /*...*/
)
