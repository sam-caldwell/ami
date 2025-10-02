package driver

type irSymbolsIndex struct {
    Schema  string               `json:"schema"`
    Package string               `json:"package"`
    Units   []irSymbolsIndexUnit `json:"units"`
}
// unit type, collectors and writer moved to separate files
