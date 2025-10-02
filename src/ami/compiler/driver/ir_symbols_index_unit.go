package driver

type irSymbolsIndexUnit struct {
    Unit    string   `json:"unit"`
    Exports []string `json:"exports,omitempty"`
    Externs []string `json:"externs,omitempty"`
}

