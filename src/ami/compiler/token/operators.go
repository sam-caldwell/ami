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
