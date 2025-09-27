package token

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
}
