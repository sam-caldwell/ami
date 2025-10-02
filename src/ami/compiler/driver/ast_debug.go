package driver

type astUnit struct {
    Schema    string       `json:"schema"`
    Package   string       `json:"package"`
    Unit      string       `json:"unit"`
    Pragmas   []astPragma  `json:"pragmas,omitempty"`
    Imports   []astImport  `json:"imports,omitempty"`
    Funcs     []astFunc    `json:"funcs,omitempty"`
    Pipelines []astPipe    `json:"pipelines,omitempty"`
}
