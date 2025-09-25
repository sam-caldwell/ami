package token

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
