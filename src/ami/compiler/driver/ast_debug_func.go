package driver

type astFunc struct {
    Name       string         `json:"name"`
    TypeParams []astTypeParam `json:"typeParams,omitempty"`
    Params     []string       `json:"params,omitempty"`
    Results    []string       `json:"results,omitempty"`
    Decorators []astDecorator `json:"decorators,omitempty"`
    Pos        *dbgPos        `json:"pos,omitempty"`
    NamePos    *dbgPos        `json:"namePos,omitempty"`
    Kind       string         `json:"kind"`
}

