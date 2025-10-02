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

// Operators maps operator lexemes to their token kinds.
var Operators = map[string]Kind{
    "=":  Assign,
    "+":  Plus,
    "-":  Minus,
    "*":  Star,
    "/":  Slash,
    "%":  Percent,
    "!":  Bang,
    "&":  BitAnd,
    "^":  BitXor,
    "|":  BitOr,
    "<<": Shl,
    ">>": Shr,

    "==": Eq,
    "!=": Ne,
    "<":  Lt,
    ">":  Gt,
    "<=": Le,
    ">=": Ge,

    "&&": And,
    "||": Or,
    "->": Arrow,
}

// Keywords is the canonical mapping from lower-case lexemes to keyword kinds.
var Keywords = map[string]Kind{
    "package":  KwPackage,
    "import":   KwImport,
    "func":     KwFunc,
    "return":   KwReturn,
    "var":      KwVar,
    "defer":    KwDefer,
    // AMI domain-reserved (pipeline semantics)
    "pipeline": KwPipeline,
    "ingress":  KwIngress,
    "egress":   KwEgress,
    "error":    KwError,
    // Primitive types
    "bool":     KwBool,
    "byte":     KwByte,
    "int":      KwInt,
    "int8":     KwInt8,
    "int16":    KwInt16,
    "int32":    KwInt32,
    "int64":    KwInt64,
    "int128":   KwInt128,
    "uint":     KwUint,
    "uint8":    KwUint8,
    "uint16":   KwUint16,
    "uint32":   KwUint32,
    "uint64":   KwUint64,
    "uint128":  KwUint128,
    "float32":  KwFloat32,
    "float64":  KwFloat64,
    "string":   KwStringTy,
    "rune":     KwRune,
    // Additional reserved words
    "append":  KwAppend,
    "break":   KwBreak,
    "case":    KwCase,
    "const":   KwConst,
    "close":   KwClose,
    "complex": KwComplex,
    "complex64": KwComplex64,
    "complex128": KwComplex128,
    "continue": KwContinue,
    "copy":    KwCopy,
    "default": KwDefault,
    "delete":  KwDelete,
    "else":    KwElse,
    "enum":    KwEnum,
    "errorevent": KwErrorEvent,
    "errorpipeline": KwErrorPipeline,
    "event":   KwEvent,
    "fallthrough": KwFallthrough,
    "false":   KwFalse,
    "float":   KwFloat,
    "for":     KwFor,
    "goto":    KwGoto,
    "if":      KwIf,
    "interface": KwInterface,
    "label":   KwLabel,
    "latest":  KwLatest,
    "len":     KwLen,
    "make":    KwMake,
    "map":     KwMap,
    "set":     KwSet,
    "slice":   KwSlice,
    "new":     KwNew,
    "nil":     KwNil,
    "panic":   KwPanic,
    "range":   KwRange,
    "real":    KwReal,
    "recover": KwRecover,
    "select":  KwSelect,
    "state":   KwState,
    "struct":  KwStruct,
    "switch":  KwSwitch,
    "true":    KwTrue,
    "type":    KwType,
    // Node keywords
    "transform": KwNodeTransform,
    "fanout":    KwNodeFanout,
    "collect":   KwNodeCollect,
    "mutable":   KwNodeMutable,
}
