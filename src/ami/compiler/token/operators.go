package token

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
