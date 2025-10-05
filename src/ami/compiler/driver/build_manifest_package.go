package driver

type bmPackage struct {
    Name          string   `json:"name"`
    Units         []bmUnit `json:"units"`
    EdgesIndex    string   `json:"edgesIndex,omitempty"`
    AsmIndex      string   `json:"asmIndex,omitempty"`
    IRIndex       string   `json:"irIndex,omitempty"`
    IRTypesIndex  string   `json:"irTypesIndex,omitempty"`
    IRSymbolsIndex string  `json:"irSymbolsIndex,omitempty"`
    WorkersLib    string   `json:"workersLib,omitempty"`
    WorkersSymbolsIndex string `json:"workersSymbolsIndex,omitempty"`
}
