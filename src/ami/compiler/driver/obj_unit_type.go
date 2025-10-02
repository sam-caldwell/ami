package driver

type objUnit struct {
    Schema    string   `json:"schema"`
    Package   string   `json:"package"`
    Unit      string   `json:"unit"`
    Functions []string `json:"functions"`
    Symbols   []objSym  `json:"symbols"`
    Relocs    []objRel  `json:"relocs"`
}

