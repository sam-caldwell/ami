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
}
