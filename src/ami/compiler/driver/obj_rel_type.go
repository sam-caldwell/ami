package driver

type objRel struct {
    Off   uint64 `json:"off"`
    Type  string `json:"type"`
    Sym   string `json:"sym"`
    Add   int64  `json:"add"`
}

